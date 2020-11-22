import re

actionEmojis = {

"created" : "\U0001F4A1",
"added" : "\U0001F4AC",
"lowered" : "\U0001F53B",
"raised" : "\U0001F53A",
"awarded" : "\U0001F3C6",
"triaged" : "\U0001F4AD",
"updated" : "\U0001F449",
"added" : "\U0001F9F3",
"changed" : "\U0000270F\U0000FE0F ",
"claimed" : "\U0001F44C",
"set" : "\U0000270F\U0000FE0F ",
"updated" : "\U0000270F\U0000FE0F ",
"reopened" : "\U0001F504",
"closed" : "\U0001F510",
"renamed" : "\U0001F449",

}

def getEmoji(txt):
	action = txt.split()[1]
	resultEmoji = ""
	for a in actionEmojis.keys():
		if a in action:
			resultEmoji = actionEmojis[a]
	return resultEmoji

def addNewlines(txt):
	return txt.replace(" from ","\nfrom ").replace(" to ","\nto ").replace(" as ","\nas ").replace(" on ","\non ")
