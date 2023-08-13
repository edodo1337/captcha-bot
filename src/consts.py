import os

from aiogram import Bot
from aiogram.filters.callback_data import CallbackData
from aiogram.fsm.state import State, StatesGroup

API_TOKEN = os.getenv("TG_API_TOKEN")
ANSWERS_COUNT = int(os.getenv("ANSWERS_COUNT", 3))
CORRECT_ANSWER_KEY = "correct_answer"
MESSAGE_IDS_LIST_KEY = "message_id"
ANSWER_TIMEOUT_SEC = int(os.getenv("ANSWER_TIMEOUT_SEC", 30))


class NewMemberState(StatesGroup):
    default = State()
    check = State()
    attempt1 = State()
    kick = State()
    approved = State()


class ChosenAnswerData(CallbackData, prefix="check-answer"):
    chosen_answer: int | None = None


BOT = Bot(token=API_TOKEN)
