#!/bin/bash

if [ -f "client" ]; then
    echo "Запуск скомпилированного клиента..."
    ./client
elif [ -f "client.go" ]; then
    echo "Запуск Go клиента..."
    go run client.go
elif [ -f "client.py" ]; then
    echo "Запуск Python клиента..."
    python3 client.py
elif [ -f "client.js" ]; then
    echo "Запуск JavaScript клиента..."
    node client.js
elif [ -f "client.class" ]; then
    echo "Запуск Java клиента..."
    java client
else
    echo "Не найден файл клиента для запуска"
    exit 1
fi