#ifndef LOG_H
#define LOG_H

#include <chrono>
#include <ctime>
#include <iomanip>
#include <sstream>
#include <string>

enum LogLevel { logInfo, logError, logWarning, logDebug, logTrace };

extern LogLevel desiredLogLevel;

inline std::string NowTime();
static std::string LevelToString(LogLevel level);

extern std::ostringstream os;

extern void setLogLevel(LogLevel level);

static std::ostringstream &Logger(LogLevel level = logInfo) {
  using namespace std::chrono;

  auto now = system_clock::now();
  auto ms = duration_cast<milliseconds>(now.time_since_epoch()) % 1000;

  auto t = std::time(nullptr);
  auto tm = *std::localtime(&t);

  os << std::put_time(&tm, "%Y-%m-%dT%H-%M-%S");
  os << '.' << std::setfill('0') << std::setw(3) << ms.count();
  os << std::put_time(&tm, "%z");
  os << " " << LevelToString(level) << " ";
  os << "VWCP"
     << " ";
  return os;
}

static std::string FlushLog() {
  std::string str = os.str();
  os.str(std::string());
  return str;
}

static std::string LevelToString(LogLevel level) {
  static const char *const buffer[] = {"[Info]", "[Error]", "[Warning]",
                                       "[Debug]", "[Trace]"};
  return buffer[level];
}

static LogLevel LevelFromString(const std::string &level) {
  if (level == "INFO")
    return logInfo;
  if (level == "ERROR")
    return logError;
  if (level == "WARNING")
    return logWarning;
  if (level == "DEBUG")
    return logDebug;
  if (level == "TRACE")
    return logTrace;

  return logInfo;
}

#define LOG(level)                                                             \
  if (level > desiredLogLevel)                                                 \
    ;                                                                          \
  else                                                                         \
    Logger(level)

#endif
