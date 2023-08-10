import asyncio
import datetime
import logging
import os
import random
from aiogram import Bot, Dispatcher, Router, types
from aiogram.types import (
    InlineKeyboardButton,
)
from aiogram.utils.keyboard import InlineKeyboardBuilder
from aiogram.filters.chat_member_updated import (
    ChatMemberUpdatedFilter,
    MEMBER,
    IS_NOT_MEMBER,
)
from aiogram.fsm.state import StatesGroup, State
from aiogram.fsm.context import FSMContext
from aiogram.enums.chat_member_status import ChatMemberStatus
from aiogram.filters.callback_data import CallbackData


logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("catpcha_bot")

API_TOKEN = os.getenv("TG_API_TOKEN")
ANSWERS_COUNT = 3
CORRECT_ANSWER_KEY = "correct_answer"
MESSAGE_ID_KEY = "message_id"
ANSWER_TIMEOUT_SEC = 120

bot = Bot(token=API_TOKEN)
router = Router()


class NewMemberState(StatesGroup):
    default = State()
    check = State()
    kick = State()
    approved = State()


class ChosenAnswerData(CallbackData, prefix="check-answer"):
    chosen_answer: int | None = None


async def background_ban_countdown(
    until_datetime: datetime.datetime,
    state: FSMContext,
    user: types.User,
    chat: types.Chat,
    captcha_message_id: str,
):
    logger.info(f"Run countdown for user_id={user.id} username={user.username}")

    while True:
        await asyncio.sleep(1)

        if datetime.datetime.now() < until_datetime:
            continue

        current_state = await state.get_state()

        if current_state is None or current_state in (
            NewMemberState.approved,
            NewMemberState.default,
        ):
            continue

        logger.info(
            f"Answer timeout, ban user_id={user.id} " f"username={user.username}"
        )
        await bot.delete_message(
            message_id=captcha_message_id,
            chat_id=chat.id,
        )
        await bot.ban_chat_member(
            chat_id=chat.id,
            user_id=user.id,
            revoke_messages=True,
        )
        return


async def is_admin(chat: types.Chat, user: types.User) -> bool:
    chat_member = await bot.get_chat_member(
        chat.id,
        user.id,
    )
    return chat_member.status in [
        ChatMemberStatus.CREATOR,
        ChatMemberStatus.ADMINISTRATOR,
    ]


@router.callback_query(ChosenAnswerData.filter())
async def process_button(
    query: types.CallbackQuery,
    callback_data: ChosenAnswerData,
    state: FSMContext,
):
    logger.info("Process answer button callback")
    state_data = await state.get_data()
    correct_answer = state_data.get(CORRECT_ANSWER_KEY)
    captcha_msg_id = state_data.get(MESSAGE_ID_KEY)

    if correct_answer is not None:
        if callback_data.chosen_answer != correct_answer:
            logger.info(
                f"Incorrect answer, ban user_id={query.from_user.id} "
                f"username={query.from_user.username}"
            )
            await bot.ban_chat_member(
                chat_id=query.message.chat.id,
                user_id=query.from_user.id,
                revoke_messages=True,
            )
        else:
            logger.info(
                f"Correct answer, approve user_id={query.from_user.id} "
                f"username={query.from_user.username}"
            )
            permissions = types.ChatPermissions(
                can_send_messages=True,
                can_send_media_messages=True,
                can_send_polls=True,
                can_send_other_messages=True,
                can_add_web_page_previews=True,
                can_change_info=False,
                can_invite_to_chat=False,
                can_pin_messages=False,
            )
            await bot.restrict_chat_member(
                query.message.chat.id,
                query.from_user.id,
                permissions=permissions,
            )

    if captcha_msg_id is not None:
        await bot.delete_message(
            message_id=captcha_msg_id, chat_id=query.message.chat.id
        )

    logger.info("Set state approved")
    await state.set_state(NewMemberState.approved)


@router.chat_member(
    ChatMemberUpdatedFilter(IS_NOT_MEMBER >> MEMBER),
)
async def on_user_joined(
    event: types.ChatMemberUpdated,
    state: FSMContext,
) -> None:
    logger.info(
        f"New user joined: "
        f"{event.old_chat_member.status} -> {event.new_chat_member.status}"
    )

    permissions = types.ChatPermissions(
        can_send_messages=False,
        can_send_media_messages=False,
        can_send_polls=False,
        can_send_other_messages=False,
        can_add_web_page_previews=False,
        can_change_info=False,
        can_invite_to_chat=False,
        can_pin_messages=False,
    )

    user_is_admin = await is_admin(event.chat, event.from_user)
    if not user_is_admin:
        logger.info("Restrict new member")
        await bot.restrict_chat_member(
            event.chat.id, event.from_user.id, permissions=permissions
        )

    x = random.randint(1, 10)
    y = random.randint(1, 10)
    answer = x + y
    equation = f"{x} + {y} = ?"
    member_name = (
        f"@{event.from_user.username}"
        if event.from_user.username
        else event.from_user.full_name
    )
    text = str(
        f"Привет <b>{member_name}</b>! "
        f"Cколько будет {equation} "
        "У тебя 2 минуты на ответ"
    )

    answer_options = [random.randint(1, 20) for _ in range(ANSWERS_COUNT - 1)]
    answer_options.append(answer)
    random.shuffle(answer_options)
    logger.info(
        f"Captcha generated for user_id={event.from_user.id} "
        f"username={event.from_user.username}"
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
            MESSAGE_ID_KEY: sent_message.message_id,
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
            captcha_message_id=sent_message.message_id,
        ),
    )


async def main() -> None:
    dp = Dispatcher()
    dp.include_router(router)

    await dp.start_polling(bot)


if __name__ == "__main__":
    asyncio.run(main())
