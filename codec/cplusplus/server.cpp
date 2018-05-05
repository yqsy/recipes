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

#include <memory>
#include <iostream>
#include <functional>
#include <thread>
#include <unordered_map>

#include <common.h>

uint16_t port = 20001;


using defer = std::shared_ptr<void>;


class Callback {
public:
    virtual ~Callback() {}

    virtual void onMessage(int sockfd, google::protobuf::Message *message) = 0;
};


template<typename T>
class CallbackT : public Callback {
    typedef std::function<void(int, T *)> MessageTCallback;

public:
    CallbackT(const MessageTCallback &callback_) : callback_(callback_) {}

    virtual void onMessage(int sockfd, google::protobuf::Message *message) {
        T *down = dynamic_cast<T *>(message);
        callback_(sockfd, down);
    }


private:
    MessageTCallback callback_;
};


struct Global {
    std::unordered_map<const google::protobuf::Descriptor *, Callback *> callbacksMap;

};

void fooQuery(int sockfd, codec::Query *query) {
    printf("%d %s\n", sockfd, query->question().c_str());

    codec::Answer answer;
    answer.set_answer("i m fine thank you, and you?");

    if (!writeAMessage(answer, sockfd)) {

        // error do nothing
        return;
    }
}

void fooEmpty(int sockfd, codec::Empty *empty) {
    printf("%d empty\n", sockfd);
}


void serveConn(int sockfd, const Global &gb) {
    defer _(nullptr, [&](...) { ::close(sockfd); });

    for (;;) {
        auto message = readAMessage(sockfd);

        if (!message) {
            printf("%d read message error\n", sockfd);
            break;
        }

        // callback
        auto it = gb.callbacksMap.find(message->GetDescriptor());
        if (it != gb.callbacksMap.end()) {
            it->second->onMessage(sockfd, &*message);
        } else {
            printf("%d gb.callbacksMap.find error\n", sockfd);
            break;
        }
    }

    printf("%d close\n", sockfd);
}

int main() {
    int listenfd = ::socket(AF_INET, SOCK_STREAM, 0);

    if (listenfd < 0) {
        panic("socket");
    }

    int yes = 1;
    if (::setsockopt(listenfd, SOL_SOCKET, SO_REUSEADDR, &yes, sizeof(yes))) {
        panic("setsocketopt");
    }

    ::signal(SIGPIPE, SIG_IGN);

    struct sockaddr_in addr;
    bzero(&addr, sizeof(addr));
    addr.sin_family = AF_INET;
    addr.sin_port = htons(port);
    addr.sin_addr.s_addr = INADDR_ANY;

    if (::bind(listenfd, reinterpret_cast<struct sockaddr *>(&addr), sizeof(addr))) {
        panic("bind");
    }

    if (::listen(listenfd, 5)) {
        panic("listen");
    }

    // 初始化回调函数
    Global gb;
    gb.callbacksMap[codec::Query::descriptor()] = new CallbackT<codec::Query>(fooQuery);
    gb.callbacksMap[codec::Empty::descriptor()] = new CallbackT<codec::Empty>(fooEmpty);

    for (;;) {
        struct sockaddr_in peerAddr;
        bzero(&peerAddr, sizeof(peerAddr));
        socklen_t addrlen = 0;
        int sockfd = ::accept(
                listenfd, reinterpret_cast<struct sockaddr *>(&peerAddr), &addrlen);

        int optval = 0;
        if (::setsockopt(sockfd, SOL_SOCKET, TCP_NODELAY,
                         &optval, static_cast<socklen_t>(sizeof optval))) {
            panic("setsocketopt");
        }

        if (sockfd < 0) {
            panic("accept");
        }

        std::thread thr(serveConn, sockfd, gb);
        thr.detach();
    }
}
