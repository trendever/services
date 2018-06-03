#!/usr/bin/python
import frida
import sys

session = frida.get_usb_device().attach("com.instagram.android")
script = session.create_script(open("stuff.js").read())
def on_message(message, data):
	print(message)
script.on('message', on_message)
script.load() 
sys.stdin.read()
session.detach()
