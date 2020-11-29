import telebot
import json
import requests
import pickle
import time
import yaml
import re

from datetime import datetime
from parseTxt import getEmoji, addNewlines


with open("config.yml") as data_file:
    config = yaml.load(data_file, Loader=yaml.BaseLoader)


chatId = config['chatId']
telegramToken = config['telegramToken']

tb = telebot.AsyncTeleBot(telegramToken)

apiToken = config['apiToken']

feedQueryUrl = config['feedQueryUrl']
phidQueryUrl = config['phidQueryUrl']

omitPattern = re.compile(r"PHID-(PROJ|USER|CMIT)-[a-z0-9]{20}")

try:
    with open('lastChrono', 'rb') as fp:
        lastChrono = pickle.load(fp) # load the pickle file
        print(lastChrono)
except (FileNotFoundError, EOFError): # error catch for if the file is not found or empty
    lastChrono = 0

feedPayload =	{
			'api.token': apiToken,
			'view': 'text',
		}


if lastChrono != 0:
	feedPayload['before'] = lastChrono

r = requests.get(feedQueryUrl, data=feedPayload)
q = json.loads(r.content)

telegramPayload = ""

ommitedMessages = 0

if q['result']:
	for i in reversed(list(q['result'])): 
		objectId = q['result'][i]['objectPHID']
		objectTxt = q['result'][i]['text']
		if omitPattern.match(objectId):
			ommitedMessages = ommitedMessages + 1 
		else:
			print(q['result'][i]['text'])
			currChrono = int(q['result'][i]['chronologicalKey'])
			if currChrono > lastChrono:
				lastChrono = currChrono
			objectTime = datetime.fromtimestamp(q['result'][i]['epoch']).strftime("%Y-%m-%d %I:%M:%S")
			phidinfo = requests.get(phidQueryUrl, data={'api.token':apiToken,'phids[0]':objectId})
			objectUri = phidinfo.json()['result'][objectId]['uri']
			print(objectUri)
			telegramPayload = getEmoji(objectTxt) + " " + addNewlines(objectTxt) + "\n\n\U0001F517 Link: " + objectUri + "\n\U0001F4C5 Kiedy: " + objectTime
			if(len(q['result']) > 80):
				time.sleep(2)
			tb.send_message(chatId, telegramPayload)
			with open('lastChrono', 'wb') as fp:
				pickle.dump(lastChrono, fp)
	if ommitedMessages > 0:
		tb.send_message(chatId, "Pominiętych wiadomości: " + str(ommitedMessages) + " - Sprawdź na https://phabricator.hskrk.pl/feed.", disable_notification=True) 

else:
	print("Brak nowych wiadomosci")
