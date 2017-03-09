#include "./log.h"
#include "./v8worker.h"

#include <iostream>
#include <thread>
#include <chrono>

int main() {
    Log(logInfo) << "hello world\n";

    // std::this_thread::sleep_for (std::chrono::seconds(1));

    Foo f = Foo(100);
    f.PrintA();

    std::cout << FlushLog() << std::endl;
    Log(logInfo) << "I'm done\n";
    std::cout << FlushLog() << std::endl;
}
