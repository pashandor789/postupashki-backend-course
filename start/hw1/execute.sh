#!/bin/bash

# Передаем все аргументы командной строки в hedgedcurl
ARGS="$@"

if [ -f "hedgedcurl" ]; then
    echo "Запуск скомпилированного hedgedcurl..."
    ./hedgedcurl $ARGS
elif [ -f "hedgedcurl.go" ]; then
    echo "Запуск Go hedgedcurl..."
    go run hedgedcurl.go $ARGS
elif [ -f "hedgedcurl.py" ]; then
    echo "Запуск Python hedgedcurl..."
    python3 hedgedcurl.py $ARGS
elif [ -f "hedgedcurl.class" ]; then
    echo "Запуск Java hedgedcurl..."
    java hedgedcurl $ARGS
else
    echo "Не найден файл hedgedcurl для запуска"
    echo "Поддерживаемые файлы: hedgedcurl.{cpp,go,py,java}"
    exit 1
fi