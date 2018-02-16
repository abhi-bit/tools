#ifndef QUEUE_H
#define QUEUE_H

#include <atomic>
#include <condition_variable>
#include <mutex>
#include <queue>
#include <thread>

template <typename T> class Queue {
private:
  std::queue<T> data_queue;
  std::mutex mut;
  std::condition_variable data_cond;
  std::atomic<std::int64_t> entry_count;

public:
  Queue() = default;
  Queue(const Queue &) = delete;
  Queue &operator=(const Queue &) = delete;

  T pop() {
    std::unique_lock<std::mutex> lk(mut);
    while (data_queue.empty()) {
      data_cond.wait(lk);
    }
    auto value = data_queue.front();
    data_queue.pop();
    entry_count--;
    return value;
  }

  void pop(T &item) {
    std::unique_lock<std::mutex> lk(mut);
    while (data_queue.empty()) {
      data_cond.wait(lk);
    }
    item = data_queue.front();
    data_queue.pop();
    entry_count--;
  }

  void push(const T &item) {
    std::unique_lock<std::mutex> lk(mut);
    data_queue.push(item);
    entry_count++;
    lk.unlock();
    data_cond.notify_one();
  }

  int64_t count() { return entry_count; }
};

#endif
