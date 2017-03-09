#ifndef __LOG_H__
#define __LOG_H__

#include <sstream>
#include <string>

enum LogLevel { logInfo, logError, logWarning, logDebug };

#ifndef LOG_MAX_LEVEL
#define LOG_MAX_LEVEL logError
#endif

#define LOG(level) \
  if (level > LOG_MAX_LEVEL) ; \
else Log(level)

#if defined(WIN32) || defined(_WIN32) || defined(__WIN32__)

#include <windows.h>

inline std::string NowTime() {
  const int MAX_LEN = 200;
  char buffer[MAX_LEN];
  if (GetTimeFormatA(LOCALE_USER_DEFAULT, 0, 0, "HH':'mm':'ss", buffer,
                     MAX_LEN) == 0)
    return "Error in NowTime()";

  char result[100] = {0};
  static DWORD first = GetTickCount();
  std::sprintf(result, "%s.%03ld", buffer,
               (long)(GetTickCount() - first) % 1000);
  return result;
}

#else

#include <sys/time.h>

inline std::string NowTime() {
  char buffer[11];
  time_t t;
  time(&t);
  tm r = {0};
  strftime(buffer, sizeof(buffer), "%X", localtime_r(&t, &r));
  struct timeval tv;
  gettimeofday(&tv, 0);
  char result[100] = {0};
  std::sprintf(result, "%s.%03ld", buffer, (long)tv.tv_usec / 1000);
  return result;
}

#endif // WIN32

inline std::string NowTime();
static std::string LevelToString(LogLevel level);

static std::ostringstream os;

static std::ostringstream& Log(LogLevel level = logInfo) {
  os << NowTime();
  os << " " << LevelToString(level) << " ";
  return os;
}

static std::string FlushLog() {
  std::string str = os.str();
  os.str(std::string());
  return str;
}

static std::string LevelToString(LogLevel level) {
  static const char *const buffer[] = {"INFO", "ERROR", "WARNING", "DEBUG"};
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

  return logInfo;
}

#endif // __LOG_H__
