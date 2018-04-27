cmake_minimum_required(VERSION 3.9)
project(asio)

set(CMAKE_CXX_STANDARD 11)

add_executable(server server.cpp)
target_link_libraries(server pthread)

add_executable(client client.cpp)
target_link_libraries(client pthread)
