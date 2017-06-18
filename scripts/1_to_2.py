#!/usr/bin/python

import sqlite3
import sys

if len(sys.argv) < 2:
  print "Usage: 1_to_2 <location of db file>"
  exit()

conn = sqlite3.connect(sys.argv[1])

conn.execute("create table category (id INTEGER PRIMARY KEY AUTOINCREMENT, owner INTEGER, name TEXT)")

conn.execute("alter table entry add column categories TEXT")

conn.execute("update entry set categories = ''")
conn.commit()
conn.close()

