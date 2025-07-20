#!/bin/bash

if [ -f "client.c" ]; then
    echo "Компиляция C клиента..."
    gcc -o client client.c
elif [ -f "client.cpp" ]; then
    echo "Компиляция C++ клиента..."
    g++ -o client client.cpp
elif [ -f "client.go" ]; then
    echo "Go не требует предварительной компиляции"
    exit 0
elif [ -f "client.py" ]; then
    echo "Python не требует компиляции"
    exit 0
elif [ -f "client.js" ]; then
    echo "JavaScript не требует компиляции"
    exit 0
elif [ -f "client.java" ]; then
    echo "Компиляция Java клиента..."
    javac client.java
else
    echo "Не найден файл клиента для компиляции"
    exit 1
fi

echo "Компиляция клиента завершена"