cmake_minimum_required(VERSION 3.9)
project(cplusplus_mutex)

set(CMAKE_CXX_STANDARD 14)

add_executable(cplusplus_mutex main.cpp)
target_link_libraries(cplusplus_mutex pthread)