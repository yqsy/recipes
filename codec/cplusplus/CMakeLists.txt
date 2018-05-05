cmake_minimum_required(VERSION 2.8)
project(cplusplus)

set(CMAKE_CXX_STANDARD 11)


add_executable(test_prototype
        ${CMAKE_HOME_DIRECTORY}/test_prototype.cpp
        ${CMAKE_HOME_DIRECTORY}/proto/query.pb.cc
        ${CMAKE_HOME_DIRECTORY}/proto/query.pb.h)

target_include_directories(test_prototype PRIVATE ${CMAKE_HOME_DIRECTORY})

target_link_libraries(test_prototype protobuf)


add_executable(server
        ${CMAKE_HOME_DIRECTORY}/server.cpp
        ${CMAKE_HOME_DIRECTORY}/common.cpp
        ${CMAKE_HOME_DIRECTORY}/proto/query.pb.cc
        ${CMAKE_HOME_DIRECTORY}/proto/query.pb.h)

target_include_directories(server PRIVATE ${CMAKE_HOME_DIRECTORY})

target_link_libraries(server protobuf pthread)

add_executable(client
        ${CMAKE_HOME_DIRECTORY}/client.cpp
        ${CMAKE_HOME_DIRECTORY}/common.cpp
        ${CMAKE_HOME_DIRECTORY}/proto/query.pb.cc
        ${CMAKE_HOME_DIRECTORY}/proto/query.pb.h)

target_include_directories(client PRIVATE ${CMAKE_HOME_DIRECTORY})

target_link_libraries(client protobuf)
