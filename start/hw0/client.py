import socket

client = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
HOST = 'localhost'
PORT = 8080

try:
    with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as client:
        client.connect((HOST, PORT))
        data = client.recv(1024)

        if data == b"OK\n":
            print("Получены корректные данные от сервера : ОК")
        else:
            print("Данные, полученные от сервера некорректны!")

except ConnectionRefusedError:
    print(f"Не удалось подключиться к серверу {HOST}:{PORT}. Убедитесь, что сервер запущен!")
except Exception as e:
    print(f"Произошла сетевая ошибка: {e}")
