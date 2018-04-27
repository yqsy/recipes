#include <iostream>
#include <boost/lockfree/queue.hpp>
#include <atomic>

#include <chrono>
#include <thread>

const int SwitchTimes = 10000000;

int main() {
    boost::lockfree::queue<int> queue(128);

    auto t1 = std::chrono::high_resolution_clock::now();

    std::thread thr([&]() {

        for (int i = 0; i < SwitchTimes; ++i) {
            queue.push(i);
        }
    });

    int consumer = 0;

    while (consumer < SwitchTimes) {
        int value;
        while (queue.pop(value)) {
            consumer++;
        }
    }


    thr.join();

    auto t2 = std::chrono::high_resolution_clock::now();

    std::chrono::duration<double, std::milli> fpMs = t2 - t1;

    auto elapsed = fpMs.count();
    std::printf("SwitchTimes:%d took:%fms speed:%.2f/s\n", SwitchTimes, elapsed,
                double(SwitchTimes) / double(elapsed / 1000));

}

