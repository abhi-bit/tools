#include <cassert>
#include <chrono>
#include <iostream>
#include <map>
#include <mutex>
#include <thread>

std::mutex assert_flags_lck;
std::map<std::string, bool> assert_flags;

bool ASSERT_SINGLE_THREAD_ENTRY(std::string context) {
  std::lock_guard<std::mutex> lock(assert_flags_lck);
  auto pos = assert_flags.find(context);
  if (pos == assert_flags.end()) {
    assert_flags[context] = false;
    std::cout << std::this_thread::get_id() << " Inserted entry for context: " << context
              << " content: " << assert_flags[context] << std::endl;
  }

  pos = assert_flags.find(context);
  std::cout << std::this_thread::get_id() << " Picking up entry for context: " << context
            << " content: " << pos->second << std::endl;
  if (!pos->second) {
    assert_flags[pos->first] = true;
    std::cout <<std::this_thread::get_id() <<  " Updating entry for context: " << pos->first
              << " content: true" << std::endl;
    return 1;
  } else {
    std::cout << std::this_thread::get_id() << " Failed try_lock, context: " << context << std::endl;
    assert(0);
    return 0;
  }
}

void ASSERT_SINGLE_THREAD_EXIT(std::string context) {
  auto pos = assert_flags.find(context);
  if (pos == assert_flags.end()) {
    std::cout << std::this_thread::get_id() << " Context: " << context << " went missing, unexpected"
              << std::endl;
    assert(0);
    return;
  }
  std::cout << std::this_thread::get_id()
            << " Resetting value for context: " << context << " to false"
            << std::endl;
  pos->second = false;
}

void counter1() {
  auto ok = ASSERT_SINGLE_THREAD_ENTRY("counter1");
  if (!ok) {
    std::cout << std::this_thread::get_id() << " Exiting counter1 as try lock failed" << std::endl;
    return;
  }
  for (int i = 0; i < 10; i++) {
    i++;
  }

  std::this_thread::sleep_for(std::chrono::milliseconds(1000));
  ASSERT_SINGLE_THREAD_EXIT("counter1");
}

void counter2() {
  auto ok = ASSERT_SINGLE_THREAD_ENTRY("counter2");
  if (!ok) {
    std::cout << std::this_thread::get_id() << " Exiting counter2 as try lock failed" << std::endl;
    return;
  }

  for (int i = 0; i < 10; i++) {
    i++;
  }
  ASSERT_SINGLE_THREAD_EXIT("counter2");
}

#define THR_COUNT 4

int main() {
  std::thread threads[THR_COUNT];

  for (int i = 0; i < THR_COUNT; i++) {
    if ((i % 2) == 0) {
      threads[i] = std::thread(counter2);
    } else {
      threads[i] = std::thread(counter1);
    }
  }

  for (auto &th : threads) {
    th.join();
  }

  return 0;
}
