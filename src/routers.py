import asyncio
import datetime

from aiogram import Router, types
from aiogram.filters.chat_member_updated import (
    IS_NOT_MEMBER,
    MEMBER,
    RESTRICTED,
    ChatMemberUpdatedFilter,
)
from aiogram.fsm.context import FSMContext
from aiogram.types import InlineKeyboardButton
from aiogram.utils.keyboard import InlineKeyboardBuilder

from consts import ANSWER_TIMEOUT_SEC
from consts import BOT as bot
from consts import (
    CORRECT_ANSWER_KEY,
    MESSAGE_IDS_LIST_KEY,
    ChosenAnswerData,
    NewMemberState,
)
from helpers import is_admin
from loggers import bot_logger as logger
from logic import (
    background_ban_countdown,
    delete_messages,
    generate_captcha,
    grant_user_permissions,
    restrict_user_permissions,
)

captcha_router = Router()


@captcha_router.callback_query(ChosenAnswerData.filter())
async def process_button(
    query: types.CallbackQuery,
    callback_data: ChosenAnswerData,
    state: FSMContext,
):
    logger_extra = {
        "user_id": query.from_user.id,
        "username": query.from_user.username,
        "fullname": query.from_user.full_name,
    }
    logger.info(
        "Process answer button callback",
        extra=logger_extra,
    )
    state_data = await state.get_data()
    current_state = await state.get_state()
    correct_answer = state_data.get(CORRECT_ANSWER_KEY)
    captcha_msg_ids = state_data.get(MESSAGE_IDS_LIST_KEY, [])

    if correct_answer is None:
        return

    if callback_data.chosen_answer != correct_answer:
        logger.info(
            f"User state={current_state}",
            extra=logger_extra,
        )
        if current_state == NewMemberState.attempt1:
            logger.info(
                "Incorrect answer, attempts exceeded, ban user",
                extra=logger_extra,
            )
            await bot.ban_chat_member(
                chat_id=query.message.chat.id,
                user_id=query.from_user.id,
                revoke_messages=True,
            )
            await state.set_state(NewMemberState.kick)
            await delete_messages(captcha_msg_ids, bot, query.message.chat)
        else:
            logger.info(
                "Incorrect answer, attempt #1 for user",
                extra=logger_extra,
            )
            reply_msg = await query.message.answer(
                "⚠️Неправильный ответ, еще 1 попытка⚠️"
            )
            await state.update_data(
                {
                    MESSAGE_IDS_LIST_KEY: [
                        *captcha_msg_ids,
                        reply_msg.message_id,
                    ],
                }
            )
            await state.set_state(NewMemberState.attempt1)
    else:
        logger.info(
            "Correct answer, approve user",
            extra=logger_extra,
        )
        await grant_user_permissions(
            bot,
            query.message.chat,
            query.from_user,
        )
        await delete_messages(captcha_msg_ids, bot, query.message.chat)

        logger.info(
            "Set state approved",
            extra=logger_extra,
        )
        await state.set_state(NewMemberState.approved)


@captcha_router.chat_member(
    ChatMemberUpdatedFilter(
        IS_NOT_MEMBER >> (MEMBER | RESTRICTED),
    ),
)
async def on_user_joined(
    event: types.ChatMemberUpdated,
    state: FSMContext,
) -> None:
    logger_extra = {
        "user_id": event.from_user.id,
        "username": event.from_user.username,
        "fullname": event.from_user.full_name,
    }
    logger.info(
        (
            "New user joined: status transition: "
            f"{event.old_chat_member.status} -> {event.new_chat_member.status}"
        ),
        extra=logger_extra,
    )

    user_is_admin = await is_admin(event.chat, event.from_user, bot)
    if user_is_admin:
        return
    else:
        logger.info(
            "Restrict new member",
            extra=logger_extra,
        )
        await restrict_user_permissions(bot, event.chat, event.from_user)

    answer_options, answer, text = generate_captcha(event.from_user)

    logger.info(
        "Captcha generated",
        extra=logger_extra,
    )

    keyboard = InlineKeyboardBuilder()
    keyboard.row(
        *[
            InlineKeyboardButton(
                text=f"{opt}",
                callback_data=ChosenAnswerData(chosen_answer=opt).pack(),
            )
            for opt in answer_options
        ]
    )

    sent_message = await event.answer(
        text=text,
        reply_markup=keyboard.as_markup(),
        parse_mode="HTML",
    )
    await state.set_state(NewMemberState.check)
    await state.set_data(
        {
            CORRECT_ANSWER_KEY: answer,
            MESSAGE_IDS_LIST_KEY: [sent_message.message_id],
        }
    )
    until_datetime = datetime.datetime.now() + datetime.timedelta(
        seconds=ANSWER_TIMEOUT_SEC
    )
    asyncio.create_task(
        background_ban_countdown(
            until_datetime=until_datetime,
            state=state,
            user=event.from_user,
            chat=event.chat,
            bot=bot,
        ),
    )
