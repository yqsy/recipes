#include <arpa/inet.h>  // uint16_t
#include <assert.h>     // assert
#include <errno.h>      // perror
#include <netinet/in.h> // sockaddr_in
#include <stdio.h>      // printf
#include <stdlib.h>     // exit
#include <strings.h>    // bzero
#include <sys/socket.h> // socket
#include <sys/types.h> // some historical (BSD) implementations required this header file
#include <unistd.h> // close
#include <signal.h>
#include <netinet/tcp.h>

#include <common.h>


const char *remoteAddr = "127.0.0.1";
uint16_t port = 20001;

using defer = std::shared_ptr<void>;

int main() {

    int sockfd = ::socket(AF_INET, SOCK_STREAM, 0);

    if (sockfd < 0) {
        panic("socket");
    }

    struct sockaddr_in addr;
    addr.sin_family = AF_INET;
    addr.sin_port = htons(port);
    addr.sin_addr.s_addr = inet_addr(remoteAddr);


    if (::connect(sockfd, reinterpret_cast<struct sockaddr *>(&addr),
                  sizeof(addr))) {
        panic("connect");
    }

    defer _(nullptr, [&](...) { ::close(sockfd); });

    codec::Query query;
    query.set_question("how are you?");

    if (!writeAMessage(query, sockfd)) {
        panic("write packet error");
    } else {
        auto message = readAMessage(sockfd);

        if (!message) {
            panic("readAMessage error");
        } else {
            codec::Answer *answer = dynamic_cast<codec::Answer *>(&*message);
            printf("%s\n", answer->answer().c_str());
        }

    }
}