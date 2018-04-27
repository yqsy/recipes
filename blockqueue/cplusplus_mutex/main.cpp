#include <queue>
#include <thread>
#include <mutex>
#include <condition_variable>

#include <chrono>
#include <cstdio>

template<typename T>
class Queue {
public:

    T pop() {
        std::unique_lock<std::mutex> mlock(mutex_);
        while (queue_.empty()) {
            cond_.wait(mlock);
        }
        auto item = queue_.front();
        queue_.pop();
        return item;
    }

    void pop(T &item) {
        std::unique_lock<std::mutex> mlock(mutex_);
        while (queue_.empty()) {
            cond_.wait(mlock);
        }
        item = queue_.front();
        queue_.pop();
    }

    void push(const T &item) {
        std::unique_lock<std::mutex> mlock(mutex_);
        queue_.push(item);
        mlock.unlock();
        cond_.notify_one();
    }

    void push(T &&item) {
        std::unique_lock<std::mutex> mlock(mutex_);
        queue_.push(std::move(item));
        mlock.unlock();
        cond_.notify_one();
    }

private:
    std::queue<T> queue_;
    std::mutex mutex_;
    std::condition_variable cond_;
};

const int SwitchTimes = 10000000;

int main(int argc, char *argv[]) {
    Queue<int> queue;

    auto t1 = std::chrono::high_resolution_clock::now();

    std::thread thr([&]() {

        for (int i = 0; i < SwitchTimes; ++i) {
            queue.push(i);
        }
    });

    for (int i = 0; i < SwitchTimes; ++i) {
        int j = queue.pop();
    }

    thr.join();

    auto t2 = std::chrono::high_resolution_clock::now();

    std::chrono::duration<double, std::milli> fp_ms = t2 - t1;

    auto elapsed = fp_ms.count();
    std::printf("SwitchTimes:%d took:%fms speed:%.2f/s\n", SwitchTimes, elapsed,
                double(SwitchTimes) / double(elapsed / 1000));
    return 0;
}
