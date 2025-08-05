#!/bin/bash

if [ -f "cryptoserver" ]; then
    echo "Запуск скомпилированного crypto сервера..."
    ./cryptoserver
elif [ -f "cryptoserver.py" ]; then
    echo "Запуск Python crypto сервера..."
    python3 cryptoserver.py
elif [ -f "cryptoserver.js" ]; then
    echo "Запуск Node.js crypto сервера..."
    node cryptoserver.js
elif [ -f "cryptoserver.go" ]; then
    echo "Запуск Go crypto сервера..."
    go run cryptoserver.go
elif [ -f "cryptoserver.class" ]; then
    echo "Запуск Java crypto сервера..."
    java cryptoserver
else
    echo "Не найден исполняемый файл crypto сервера"
    echo "Убедитесь что файл скомпилирован или существует cryptoserver.{py,js,go}"
    exit 1
fi