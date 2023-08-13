import logging

logging.basicConfig(
    level=logging.INFO,
    # format=(
    #     "%(asctime)s - %(message)s"
    # ),
    # datefmt="%Y-%m-%d %H:%M:%S",
)  # global logging level


bot_logger = logging.getLogger("captcha-bot")
bot_logger.setLevel(logging.INFO)
handler = logging.StreamHandler()

log_format = (
    "%(asctime)s - user_id: %(user_id)s - "
    "username: %(username)s - %(levelname)s - %(message)s"
)
formatter = logging.Formatter(log_format, datefmt="%Y-%m-%d %H:%M:%S")
handler.setFormatter(formatter)
bot_logger.propagate = False
bot_logger.addHandler(handler)
