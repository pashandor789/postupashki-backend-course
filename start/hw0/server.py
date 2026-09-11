import socket


def main():
    server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
    server_socket.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)  # Разрешаем сразу переиспользовать сокет

    server_address = ('localhost', 8080)
    server_socket.bind(server_address)

    server_socket.listen(5)  # Начинаем слушать подключения
    print(f"Сервер запущен на {server_address[0]}:{server_address[1]}")

    try:
        while True:
            client_socket, client_address = server_socket.accept()
            print(f"Подключен клиент: {client_address}")

            try:
                client_socket.sendall(b"OK\n")
                print(f"Отправлено 'OK\\n' клиенту {client_address}")
            finally:
                client_socket.close()
                print(f"Соединение с {client_address} закрыто")

    except Exception as e:
        print(f"Ошибка: {e}")
    finally:
        server_socket.close()
        print("Сервер завершил работу")


if __name__ == "__main__":
    main()