#include "./log.h"
#include "./v8worker.h"

#include <iostream>
#include <thread>
#include <chrono>

#include <sstream>

std::ostringstream os;
LogLevel desiredLogLevel = LogLevel(0);
void setLogLevel(LogLevel level) { desiredLogLevel = level; }

int main() {
    LOG(logInfo) << "hello world\n";

    // std::this_thread::sleep_for (std::chrono::seconds(1));

    Foo f = Foo(100);
    f.PrintA();

    std::cout << FlushLog() << std::endl;
    LOG(logInfo) << "I'm done\n";
    std::cout << FlushLog() << std::endl;
}
