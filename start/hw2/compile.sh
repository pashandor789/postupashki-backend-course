#!/bin/bash

if [ -f "cryptoserver.cpp" ]; then
    echo "Компиляция C++ crypto сервера..."
    g++ -o cryptoserver cryptoserver.cpp -lcurl -lpthread -std=c++11
elif [ -f "cryptoserver.go" ]; then
    echo "Компиляция Go crypto сервера..."
    go build -o cryptoserver cryptoserver.go
elif [ -f "cryptoserver.py" ]; then
    echo "Python не требует компиляции"
    exit 0
elif [ -f "cryptoserver.js" ]; then
    echo "JavaScript не требует компиляции"
    exit 0
elif [ -f "cryptoserver.java" ]; then
    echo "Компиляция Java crypto сервера..."
    javac cryptoserver.java
else
    echo "Не найден файл cryptoserver для компиляции"
    echo "Поддерживаемые файлы: cryptoserver.{cpp,go,py,js,java}"
    exit 1
fi

echo "Компиляция cryptoserver завершена"