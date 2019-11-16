# Listserialports

When working with hardware devices with usb-serial interface, it is sometimes quite frustrating when usb ports enumerate into hard to remember names. Or it is possible to start more than one program that access same serial ports and cause "*funny*" effects.

So I made this library for listing and checking serial device availability on program start.

Check lsserials as example/standalone command line tool.


## notice

This works only on linux

This library is only part of my other project.

Later, there will be some features found from setserial utility (if my project requires)
https://github.com/brgl/busybox/blob/master/miscutils/setserial.c
like checking real uart status
