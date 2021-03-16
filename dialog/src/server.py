#!/usr/bin/env python
import os

from chatterbot import ChatBot
from chatterbot.trainers import ChatterBotCorpusTrainer
from flask import Flask, request, jsonify

server = Flask(__name__)

english_bot = ChatBot(
    'Chatterbot',
    storage_adapter='chatterbot.storage.SQLStorageAdapter',
    database_uri='mysql+pymysql://root@localhost/dialog_history',
)
trainer = ChatterBotCorpusTrainer(english_bot)
trainer.train('chatterbot-corpus.chatterbot_corpus.data.english')


@server.route('/converse', methods=['POST'])
def converse():
    user_text = request.json['message']
    return jsonify(
        {
            "reply": str(english_bot.get_response(user_text)),
        },
    )


if __name__ == "__main__":
    server.run(debug=(os.getenv('DEBUG_SERVER', False)))
