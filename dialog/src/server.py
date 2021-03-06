#!/usr/bin/env python
from chatterbot import ChatBot
from chatterbot.trainers import ChatterBotCorpusTrainer
from flask import Flask, request, jsonify

server = Flask(__name__)

english_bot = ChatBot(
    'Chatterbot',
    storage_adapter='chatterbot.storage.SQLStorageAdapter',
)
trainer = ChatterBotCorpusTrainer(english_bot)
trainer.train('chatterbot.corpus.english')


@server.route('/')
def index():
    return 'Looking to chat? Post a JSON with a \'message\' key to \'/converse\'.'


@server.route('/converse', methods=['POST'])
def converse():
    user_text = request.json['message']
    return jsonify(
        {
            "reply": str(english_bot.get_response(user_text)),
        },
    )


if __name__ == "__main__":
    server.run()
