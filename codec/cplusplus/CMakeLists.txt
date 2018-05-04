cmake_minimum_required(VERSION 3.10)
project(cplusplus)

set(CMAKE_CXX_STANDARD 11)


add_executable(cplusplus
        ${CMAKE_HOME_DIRECTORY}/main.cpp
        ${CMAKE_HOME_DIRECTORY}/proto/query.pb.cc
        ${CMAKE_HOME_DIRECTORY}/proto/query.pb.h
        )
target_include_directories(cplusplus PRIVATE ${CMAKE_HOME_DIRECTORY})

target_link_libraries(cplusplus protobuf)
