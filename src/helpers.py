from aiogram import Bot, types
from aiogram.enums.chat_member_status import ChatMemberStatus


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


async def is_admin(
    chat: types.Chat,
    user: types.User,
    bot: Bot,
) -> bool:
    chat_member = await bot.get_chat_member(
        chat.id,
        user.id,
    )
    return chat_member.status in [
        ChatMemberStatus.CREATOR,
        ChatMemberStatus.ADMINISTRATOR,
    ]
