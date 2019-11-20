

from heedy import App

c = App("gV9A4/J0eS4nLGAnzYyn")
s = c.listObjects()[1]
print(s.length())
print(s.append("hi"))
print(s.length())
print(s())
print(s.delete(i=0))
print(s.length())
print(s())
