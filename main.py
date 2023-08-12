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
    RESTRICTED,
)
from aiogram.fsm.state import StatesGroup, State
from aiogram.fsm.context import FSMContext
from aiogram.enums.chat_member_status import ChatMemberStatus
from aiogram.filters.callback_data import CallbackData


logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("catpcha_bot")

# -------------Constants------------- #
API_TOKEN = os.getenv("TG_API_TOKEN")
ANSWERS_COUNT = int(os.getenv("ANSWERS_COUNT", 3))
CORRECT_ANSWER_KEY = "correct_answer"
MESSAGE_IDS_LIST_KEY = "message_id"
ANSWER_TIMEOUT_SEC = int(os.getenv("ANSWER_TIMEOUT_SEC", 30))

bot = Bot(token=API_TOKEN)
router = Router()


class NewMemberState(StatesGroup):
    default = State()
    check = State()
    attempt1 = State()
    kick = State()
    approved = State()


class ChosenAnswerData(CallbackData, prefix="check-answer"):
    chosen_answer: int | None = None


def display_countdown_description(countdown: int) -> str:
    def _get_minute_ending(number: int) -> str:
        if 10 <= number % 100 <= 20:
            return "минут"
        elif number % 10 == 1:
            return "минута"
        elif 2 <= number % 10 <= 4:
            return "минуты"
        else:
            return "минут"

    def _get_second_ending(number: int) -> str:
        if 10 <= number % 100 <= 20:
            return "секунд"
        elif number % 10 == 1:
            return "секунда"
        elif 2 <= number % 10 <= 4:
            return "секунды"
        else:
            return "секунд"

    minutes = countdown // 60
    seconds = countdown % 60

    result = ""
    if minutes:
        result = f"{result}{minutes} {_get_minute_ending(minutes)} "

    if seconds:
        result = f"{result}{seconds} {_get_second_ending(seconds)} "

    return result


async def background_ban_countdown(
    until_datetime: datetime.datetime,
    state: FSMContext,
    user: types.User,
    chat: types.Chat,
):
    logger.info(
        f"Run countdown for user_id={user.id} username={user.username}",
    )
    state_data = await state.get_data()
    captcha_msg_ids = state_data.get(MESSAGE_IDS_LIST_KEY, [])

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
        await bot.ban_chat_member(
            chat_id=chat.id,
            user_id=user.id,
            revoke_messages=True,
        )
        for msg_id in captcha_msg_ids:
            try:
                await bot.delete_message(
                    message_id=msg_id,
                    chat_id=chat.id,
                )
            finally:
                pass

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


# -------------Bot routers and callbacks------------- #
@router.callback_query(ChosenAnswerData.filter())
async def process_button(
    query: types.CallbackQuery,
    callback_data: ChosenAnswerData,
    state: FSMContext,
):
    logger.info("Process answer button callback")
    state_data = await state.get_data()
    correct_answer = state_data.get(CORRECT_ANSWER_KEY)
    captcha_msg_ids = state_data.get(MESSAGE_IDS_LIST_KEY, [])

    if correct_answer is None:
        return

    if callback_data.chosen_answer != correct_answer:
        if state == NewMemberState.attempt1:
            logger.info(
                f"Incorrect answer, attempts exceeded, "
                f"ban user_id={query.from_user.id} "
                f"username={query.from_user.username}"
            )
            await bot.ban_chat_member(
                chat_id=query.message.chat.id,
                user_id=query.from_user.id,
                revoke_messages=True,
            )
            await state.set_state(NewMemberState.kick)
            if isinstance(captcha_msg_ids, list):
                for msg_id in captcha_msg_ids:
                    try:
                        await bot.delete_message(
                            message_id=msg_id, chat_id=query.message.chat.id
                        )
                    finally:
                        pass
        else:
            logger.info(
                f"Incorrect answer, attempt #1 "
                f"for user_id={query.from_user.id} "
                f"username={query.from_user.username}"
            )
            await query.answer("Неправильный ответ, у вас еще 1 попытка")
            await state.set_state(NewMemberState.attempt1)
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
            can_change_info=True,
            can_invite_to_chat=True,
            can_pin_messages=True,
        )
        await bot.restrict_chat_member(
            query.message.chat.id,
            query.from_user.id,
            permissions=permissions,
        )
        if isinstance(captcha_msg_ids, list):
            for msg_id in captcha_msg_ids:
                try:
                    await bot.delete_message(
                        message_id=msg_id, chat_id=query.message.chat.id
                    )
                finally:
                    pass

        logger.info("Set state approved")
        await state.set_state(NewMemberState.approved)


@router.chat_member(
    ChatMemberUpdatedFilter(
        IS_NOT_MEMBER >> (MEMBER | RESTRICTED),
    ),
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
        can_invite_to_chat=False,
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
        f"У тебя {display_countdown_description(ANSWER_TIMEOUT_SEC)} на ответ"
    )

    answer_options = [random.randint(1, 20) for _ in range(ANSWERS_COUNT - 1)]
    answer_options = [i if i != answer else i + 1 for i in answer_options]
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
        ),
    )


async def main() -> None:
    dp = Dispatcher()
    dp.include_router(router)

    await dp.start_polling(bot)


if __name__ == "__main__":
    asyncio.run(main())
