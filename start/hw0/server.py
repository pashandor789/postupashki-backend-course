import socket

HOST = 'localhost'
PORT = 8080

with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as server:
    server.setsockopt(socket.SOL_SOCKET, socket.SO_REUSEADDR, 1)
    server.bind((HOST, PORT))
    server.listen(5)
    print("Сервер запущен и ждет подключений...")

    try:
        while True:
            client, address = server.accept()
            with client:
                client.sendall(b"OK\n")
    except KeyboardInterrupt:
        print("\nСервер остановлен пользователем.")
        