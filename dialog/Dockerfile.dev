FROM python:3.8

WORKDIR /app

COPY requirements.txt .

RUN pip install -r requirements.txt

COPY src/ .

RUN pip install -e chatterbot/

RUN pip install -e chatterbot-corpus/

RUN python -m spacy download en

CMD [ "python", "./server.py" ]