import asyncio
import datetime
import logging
import random

from aiogram import Bot, types
from aiogram.fsm.context import FSMContext

from consts import (
    ANSWER_TIMEOUT_SEC,
    ANSWERS_COUNT,
    MESSAGE_IDS_LIST_KEY,
    NewMemberState,
)
from helpers import display_countdown_description
from loggers import bot_logger as logger

default_logger = logging.getLogger(__name__)


async def delete_messages(
    captcha_msg_ids: list[str],
    bot: Bot,
    chat: types.Chat,
):
    for msg_id in captcha_msg_ids:
        try:
            await bot.delete_message(
                message_id=msg_id,
                chat_id=chat.id,
            )
        except Exception as ex:
            default_logger.exception(ex)


async def background_ban_countdown(
    until_datetime: datetime.datetime,
    state: FSMContext,
    user: types.User,
    chat: types.Chat,
    bot: Bot,
):
    logger_extra = {
        "user_id": user.id,
        "username": chat.username,
    }
    logger.info(
        "Run ban countdown",
        extra=logger_extra,
    )
    state_data = await state.get_data()
    captcha_msg_ids = state_data.get(MESSAGE_IDS_LIST_KEY, [])

    while True:
        await asyncio.sleep(1)

        if datetime.datetime.now() < until_datetime:
            continue

        current_state = await state.get_state()

        if current_state in (
            NewMemberState.approved,
            NewMemberState.default,
        ):
            continue

        if current_state is None or current_state in (
            NewMemberState.approved,
            NewMemberState.kick,
        ):
            return

        logger.info(
            "Answer timeout, ban user",
            extra=logger_extra,
        )
        await bot.ban_chat_member(
            chat_id=chat.id,
            user_id=user.id,
            revoke_messages=True,
        )
        await delete_messages(captcha_msg_ids, bot, chat)

        return


def generate_captcha(
    user: types.User,
) -> tuple[str, str, str]:
    x = random.randint(1, 10)
    y = random.randint(1, 10)
    correct_answer = x + y
    equation = f"{x} + {y} = ?"
    member_name = f"@{user.username}" if user.username else user.full_name
    text = str(
        f"Привет <b>{member_name}</b>! "
        f"Cколько будет {equation} "
        f"У тебя {display_countdown_description(ANSWER_TIMEOUT_SEC)} на ответ"
    )

    answer_options = [random.randint(1, 20) for _ in range(ANSWERS_COUNT - 1)]
    answer_options = [i if i != correct_answer else i + 1 for i in answer_options]
    answer_options.append(correct_answer)
    random.shuffle(answer_options)

    return answer_options, correct_answer, text


async def restrict_user_permissions(
    bot: Bot,
    chat: types.Chat,
    user: types.User,
):
    permissions = types.ChatPermissions(
        can_send_messages=False,
        can_invite_to_chat=False,
    )
    await bot.restrict_chat_member(
        chat.id,
        user.id,
        permissions=permissions,
    )


async def grant_user_permissions(
    bot: Bot,
    chat: types.Chat,
    user: types.User,
):
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
        chat.id,
        user.id,
        permissions=permissions,
    )
