import telebot
import json
import requests
import pickle
import time
import yaml

from parseTxt import getEmoji, addNewlines


with open("config.yml") as data_file:
    config = yaml.load(data_file, Loader=yaml.BaseLoader)


chatId = config['chatId']
telegramToken = config['telegramToken']

tb = telebot.AsyncTeleBot(telegramToken)

apiToken = config['apiToken']

feedQueryUrl = config['feedQueryUrl']
phidQueryUrl = config['phidQueryUrl']

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



for i in reversed(list(q['result'])): 
	print(q['result'][i]['text'])
	currChrono = q['result'][i]['chronologicalKey'] 
	if currChrono > lastChrono:
		lastChrono = currChrono
	objectId = q['result'][i]['objectPHID']
	objectTxt = q['result'][i]['text']
	phidinfo = requests.get(phidQueryUrl, data={'api.token':apiToken,'phids[0]':objectId})
	objectUri = phidinfo.json()['result'][objectId]['uri']
	print(objectUri)
	telegramPayload = getEmoji(objectTxt) + " " + addNewlines(objectTxt) + "\n\n\U0001F517 Link: " + objectUri
	if(len(q['result']) > 100):
		time.sleep(0.5)
	tb.send_message(chatId, telegramPayload)
	with open('lastChrono', 'wb') as fp:
		pickle.dump(lastChrono, fp)
