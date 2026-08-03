#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <unistd.h>
#include <arpa/inet.h>
#include <sys/socket.h>
#include <netinet/in.h>

#define SERVER_PORT 8080
#define BACKLOG 10
#define SERVER_RESPONSE "OK\n"

int main()
{
    int server_socket = socket(AF_INET, SOCK_STREAM, 0);
    if (server_socket == -1)
    {
        perror("socket");
        return 1;
    }

    struct sockaddr_in server_addr;
    memset(&server_addr, 0, sizeof(server_addr));
    server_addr.sin_family = AF_INET;
    server_addr.sin_port = htons(SERVER_PORT);
    server_addr.sin_addr.s_addr = htonl(INADDR_ANY);

    if (bind(server_socket, (struct sockaddr *)&server_addr, sizeof(server_addr)) == -1)
    {
        perror("bind");
        close(server_socket);
        return 1;
    }

    if (listen(server_socket, BACKLOG) == -1)
    {
        perror("listen");
        close(server_socket);
        return 1;
    }

    while(1)
    {
        struct sockaddr_in client_addr;
        memset(&client_addr, 0, sizeof(client_addr));
        socklen_t client_addr_len = sizeof(client_addr);

        int client_socket = accept(server_socket, (struct sockaddr*)&client_addr, &client_addr_len);
        if (client_socket == -1)
        {
            perror("accept");
            continue;
        }
        send(client_socket, SERVER_RESPONSE, strlen(SERVER_RESPONSE), 0);
        close(client_socket);
    }

    return 0;
}
