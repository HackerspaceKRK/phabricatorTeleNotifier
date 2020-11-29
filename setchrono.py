import pickle
import sys

lastChrono = sys.argv[1]
print("Ustawiam - " + lastChrono)

with open('lastChrono', 'wb') as fp:
    pickle.dump(int(lastChrono), fp)

