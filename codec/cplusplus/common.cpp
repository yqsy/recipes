#include <common.h>

#include <stdio.h>      // printf
#include <stdlib.h>     // exit
#include <sys/socket.h> // socket
#include <unistd.h> // close
#include <errno.h>
#include <arpa/inet.h>

#include <proto/query.pb.h>

const int MaxLen = 16384;

void panic(const char *str) {
    perror(str);
    exit(1);
}

int readN(int sockfd, void *buf, int length) {
    int nread = 0;
    while (nread < length) {
        ssize_t nr = ::read(sockfd, static_cast<char *>(buf) + nread, length - nread);
        if (nr > 0) {
            nread += static_cast<int>(nr);
        } else if (nr == 0) {
            break; // EOF
        } else if (errno != EINTR) {
            panic("read");
        }
    }

    return nread;
}

int writeN(int sockfd, const void *buf, int length) {
    int written = 0;
    while (written < length) {
        ssize_t nw = ::write(sockfd, static_cast<const char *>(buf) + written, length - written);
        if (nw > 0) {
            written += static_cast<int>(nw);
        } else if (nw == 0) {
            break;  // EOF
        } else if (errno != EINTR) {
            panic("read");
        }
    }
    return written;
}


google::protobuf::Message *createMessage(const std::string &typeName) {
    google::protobuf::Message *message = nullptr;

    auto descriptor = google::protobuf::DescriptorPool::generated_pool()->FindMessageTypeByName(typeName);
    if (descriptor) {
        auto prototype = google::protobuf::MessageFactory::generated_factory()->GetPrototype(descriptor);
        if (prototype) {
            message = prototype->New();
        }
    }
    return message;
}


std::shared_ptr<google::protobuf::Message> readAMessage(int sockfd) {
    std::shared_ptr<Packet> packet = std::make_shared<Packet>();
    bzero(&*packet, sizeof(Packet));

    char twoLen[8] = {};
    int rn = readN(sockfd, twoLen, 8);
    if (rn != 8) {
        return nullptr;
    }

    packet->len = *reinterpret_cast<int *>(&twoLen[0]);
    packet->len = ntohl(packet->len);

    packet->nameLen = *reinterpret_cast<int *>(&twoLen[4]);
    packet->nameLen = ntohl(packet->nameLen);

    if (packet->len > MaxLen) {
        return nullptr;
    };

    packet->typeName = new char[packet->nameLen];
    rn = readN(sockfd, packet->typeName, packet->nameLen);
    if (rn != packet->nameLen) {
        return nullptr;
    }

    auto protobufLen = packet->len - sizeof(packet->nameLen) - packet->nameLen - sizeof(packet->checkSum);
    packet->protobufData = new char[protobufLen];
    rn = readN(sockfd, packet->protobufData, protobufLen);
    if (rn != protobufLen) {
        return nullptr;
    }

    rn = readN(sockfd, &packet->checkSum, sizeof(packet->checkSum));
    packet->checkSum = ntohl(packet->checkSum);
    if (rn != sizeof(packet->checkSum)) {
        return nullptr;
    }

    // 缺少buffer组件,所以这里就不计算checksum了
    // buffer的话和非阻塞的做法一致,peek头部4字节,然后read full了之后再做处理

    std::shared_ptr<google::protobuf::Message> smart;

    // prototype
    google::protobuf::Message *message = createMessage(packet->typeName);
    smart.reset(message);

    if (!message) {
        return nullptr;
    }

    // deserialize
    int32_t dataLen = packet->len - sizeof(int32_t) * 2 - packet->nameLen;
    if (!message->ParseFromArray(packet->protobufData, dataLen)) {
        return nullptr;
    } else {
        return smart;
    }
}

bool writeAMessage(const google::protobuf::Message &message, int sockfd) {

    auto typeName = message.GetTypeName() + "\0";
    int32_t nameLen = typeName.length() + 1 /*fuck c style string*/;
    int32_t protobufLen = message.ByteSize();
    int32_t checkSum = 0;

    int32_t len = sizeof(nameLen) + nameLen + protobufLen + sizeof(checkSum);

    int32_t bufLen = len + sizeof(len);

    char *p = new char[bufLen];
    auto begin = p;

    *reinterpret_cast<int32_t *>(p) = htonl(len);
    p += sizeof(int32_t);
    *reinterpret_cast<int32_t *>(p) = htonl(nameLen);
    p += sizeof(int32_t);
    memcpy(p, typeName.c_str(), nameLen);
    p += nameLen;
    auto end = message.SerializeWithCachedSizesToArray(reinterpret_cast<uint8_t *>(p));

    if (reinterpret_cast<char *>(end) - p != protobufLen) {
        return false;
    }

    *reinterpret_cast<int32_t *>(end) = htonl(checkSum);
    end += sizeof(int32_t);

    int wn = writeN(sockfd, begin, bufLen);

    return wn == bufLen;
}
