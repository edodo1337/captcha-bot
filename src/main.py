import asyncio
import logging

from aiogram import Dispatcher

from consts import BOT as bot
from routers import captcha_router

logger = logging.getLogger(__name__)
logging.basicConfig(
    level=logging.INFO,
    format=("%(asctime)s - %(message)s"),
    datefmt="%Y-%m-%d %H:%M:%S",
)


async def main() -> None:
    logger.info("Bot started")
    dp = Dispatcher()
    dp.include_router(captcha_router)

    await bot.delete_webhook(drop_pending_updates=True)
    await dp.start_polling(bot)


if __name__ == "__main__":
    asyncio.run(main())
