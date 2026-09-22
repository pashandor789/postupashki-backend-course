#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <arpa/inet.h>
#include <sys/socket.h>
#include <netinet/in.h>

#define SERVER_PORT 8080
#define SERVER_IP "127.0.0.1"
#define BUFFER_SIZE 16
#define EXPECTED_MESSAGE "OK\n"

int main()
{
    int client_socket = socket(AF_INET, SOCK_STREAM, 0);
    if (client_socket == -1)
    {
        perror("socket");
        return 1;
    }

    struct sockaddr_in server_addr;
    memset(&server_addr, 0, sizeof(server_addr));
    server_addr.sin_family = AF_INET;
    server_addr.sin_port = htons(SERVER_PORT);
    server_addr.sin_addr.s_addr = inet_addr(SERVER_IP);

    if (connect(client_socket, (struct sockaddr *)&server_addr,  sizeof(server_addr)) == -1)
    {
        perror("connect");
        close(client_socket);
        return 1;
    }

    char buffer[BUFFER_SIZE];
    int bytes = recv(client_socket, buffer, BUFFER_SIZE - 1, 0);

    if (bytes <= -1)
    {
        perror("recv");
        close(client_socket);
        return 1;
    }

    buffer[bytes] = '\0';
    if (strcmp(buffer, EXPECTED_MESSAGE) != 0)
    {
        close(client_socket);
        return 1;
    }
    close(client_socket);
    return 0;
}
