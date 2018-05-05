#ifndef CPLUSPLUS_COMMON_H
#define CPLUSPLUS_COMMON_H


#include <zconf.h>

#include <string>
#include <memory>

#include <proto/query.pb.h>

void panic(const char *str);

int readN(int sockfd, void *buf, int length);

int writeN(int sockfd, const void *buf, int length);

struct Packet {
    int32_t len;
    int32_t nameLen;
    char *typeName; // with \0x00 in the end
    char *protobufData; // pure protobuf data
    int32_t checkSum;

    ~Packet() {
        if (typeName) {
            delete typeName;
        }

        if (protobufData) {
            delete protobufData;
        }
    }
};

std::shared_ptr<google::protobuf::Message> readAMessage(int sockfd);

bool writeAMessage(const google::protobuf::Message &message, int sockfd);

google::protobuf::Message *createMessage(const std::string &typeName);

#endif //CPLUSPLUS_COMMON_H
