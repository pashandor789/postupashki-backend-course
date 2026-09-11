import socket


def recv_until_eof(sock, bufsize=1024):
    chunks = []
    while True:
        chunk = sock.recv(bufsize)
        if not chunk:  # сервер закрыл соединение
            return b"".join(chunks)

        chunks.append(chunk)


def main():
    client_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)

    try:
        server_address = ('localhost', 8080)
        client_socket.connect(server_address)
        response = recv_until_eof(client_socket)  # Читает данные из сокета до закрытия соединения (EOF)

        if response == b"OK\n":
            print("Получен ожидаемый ответ 'OK\\n'")
        else:
            print(f"Получен неожиданный ответ '{response}'")

    except Exception as e:
        print(f"Ошибка: {e}")
    finally:
        client_socket.close()
        print("Соединение закрыто")


if __name__ == "__main__":
    main()