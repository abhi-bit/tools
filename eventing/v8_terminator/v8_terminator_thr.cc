#include <atomic>
#include <cassert>
#include <chrono>
#include <cstdio>
#include <cstring>
#include <fstream>
#include <iostream>
#include <map>
#include <streambuf>
#include <string>
#include <thread>
#include <unistd.h>
#include <vector>

#include <include/libplatform/libplatform.h>
#include <include/v8-debug.h>
#include <include/v8.h>

using namespace std;
using namespace v8;

Isolate *isolate;
Persistent<Context> context_;
Persistent<Function> debug_user_request;
volatile std::atomic<bool> execute_flag;
std::chrono::high_resolution_clock::time_point start_time;

const string currentDateTime() {
  time_t now = time(0);
  struct tm tstruct;
  char buf[80];
  tstruct = *localtime(&now);
  strftime(buf, sizeof(buf), "%Y-%m-%d.%X", &tstruct);

  return buf;
}

int AsciiToUtf16(const char *input_buffer, uint16_t *output_buffer) {
  int i;
  for (i = 0; input_buffer[i] != '\0'; ++i) {
    // ASCII does not use chars > 127, but be careful anyway.
    output_buffer[i] = static_cast<unsigned char>(input_buffer[i]);
  }
  output_buffer[i] = 0;
  return i;
}

const char *ToCString(const String::Utf8Value &value) {
  return *value ? *value : "<string conversion failed>";
}

const char *ToJson(Isolate *isolate, Handle<Value> object) {
  HandleScope handle_scope(isolate);

  Local<Context> context = isolate->GetCurrentContext();
  Local<Object> global = context->Global();

  Local<Object> JSON =
      global->Get(String::NewFromUtf8(isolate, "JSON"))->ToObject();
  Local<Function> JSON_stringify = Local<Function>::Cast(
      JSON->Get(String::NewFromUtf8(isolate, "stringify")));

  Local<Value> result;
  Local<Value> args[1];
  args[0] = {object};
  result = JSON_stringify->Call(context->Global(), 1, args);
  String::Utf8Value str(result->ToString());
  return ToCString(str);
}

string ObjectToString(Local<Value> value) {
  String::Utf8Value utf8_value(value);
  return string(*utf8_value);
}

string ToString(Isolate *isolate, Handle<Value> object) {
  HandleScope handle_scope(isolate);

  Local<Context> context = isolate->GetCurrentContext();
  Local<Object> global = context->Global();

  Local<Object> JSON =
      global->Get(String::NewFromUtf8(isolate, "JSON"))->ToObject();
  Local<Function> JSON_stringify = Local<Function>::Cast(
      JSON->Get(String::NewFromUtf8(isolate, "stringify")));

  Local<Value> result;
  Local<Value> args[1];
  args[0] = {object};
  result = JSON_stringify->Call(context->Global(), 1, args);
  return ObjectToString(result);
}

void Print(const FunctionCallbackInfo<Value> &args) {
  bool first = true;
  for (int i = 0; i < args.Length(); i++) {
    HandleScope handle_scope(args.GetIsolate());
    if (first) {
      first = false;
    } else {
      printf(" ");
    }
    String::Utf8Value str(args[i]);
    const char *cstr = ToJson(args.GetIsolate(), args[i]);
    printf("%s", cstr);
  }
  printf("\n");
  fflush(stdout);
}

class ArrayBufferAllocator : public v8::ArrayBuffer::Allocator {
public:
  virtual void *Allocate(size_t length) {
    void *data = AllocateUninitialized(length);
    return data == NULL ? data : memset(data, 0, length);
  }
  virtual void *AllocateUninitialized(size_t length) { return malloc(length); }
  virtual void Free(void *data, size_t) { free(data); }
};

void ProcessDebugUserRequest(string request) {
  Locker locker(isolate);
  Isolate::Scope isolate_Scope(isolate);
  HandleScope handle_Scope(isolate);

  Local<Context> context = Local<Context>::New(isolate, context_);
  Context::Scope context_scope(context);

  Handle<Value> args[1];
  args[0] = JSON::Parse(String::NewFromUtf8(isolate, request.c_str()));
  cout << currentDateTime() << "" << __FUNCTION__ << request << " " << endl;
  Local<Function> handle_user_req_fun =
      Local<Function>::New(isolate, debug_user_request);

  typedef std::chrono::high_resolution_clock Time;
  typedef std::chrono::nanoseconds ns;
  typedef std::chrono::duration<float> fsec;

  execute_flag = true;
  start_time = Time::now();
  auto t0 = Time::now();
  handle_user_req_fun->Call(context->Global(), 1, args);
  auto t1 = Time::now();
  fsec fs = t1 - t0;
  execute_flag = false;

  ns d = std::chrono::duration_cast<ns>(fs);

  std::cout << "Took: " << fs.count() << "s" << std::endl;
  std::cout << "Took: " << d.count() << "ns" << std::endl;
}

void ProcessRequest(int iter_count) {
  string prefix(
      "{\"type\": \"json\", \"client\": \"Chrome Canary\", \"counter\":");
  for (int i = 0; i < iter_count; i++) {
    string request;
    request.append(prefix);
    request.append(to_string(i));
    request.append("}");
    ProcessDebugUserRequest(request);
  }
}

int max_task_duration = 1000 * 1000 * 200;

void TerminateTask() {
  while (true) {
    std::this_thread::sleep_for(std::chrono::milliseconds(100));
    if (execute_flag) {
      std::chrono::high_resolution_clock::time_point t1 =
          std::chrono::high_resolution_clock::now();

      std::chrono::duration<float> fs = t1 - start_time;
      typedef std::chrono::nanoseconds ns;
      ns d = std::chrono::duration_cast<ns>(fs);

      std::cout << "time taken: " << fs.count() << std::endl;
      std::cout << "time taken: " << d.count() << std::endl;

      if (d.count() > max_task_duration) {
        std::cout << "Terminating execution" << std::endl;
        V8::TerminateExecution(isolate);
      }
    }
  }
}

int main(int argc, char *argv[]) {
  std::cout.setf(std::ios_base::unitbuf);

  V8::InitializeICU();
  V8::InitializeExternalStartupData(argv[0]);
  Platform *platform = platform::CreateDefaultPlatform();
  V8::InitializePlatform(platform);
  V8::Initialize();

  ArrayBufferAllocator allocator;
  Isolate::CreateParams create_params;
  create_params.array_buffer_allocator = &allocator;
  isolate = Isolate::New(create_params);
  {
    Isolate::Scope isolate_scope(isolate);
    HandleScope handle_scope(isolate);
    Local<ObjectTemplate> global = ObjectTemplate::New(isolate);

    global->Set(String::NewFromUtf8(isolate, "log"),
                FunctionTemplate::New(isolate, Print));

    Local<Context> context = Context::New(isolate, NULL, global);
    context_.Reset(isolate, context);
    Context::Scope context_scope(context);

    ifstream file_name(argv[1]);
    string src((istreambuf_iterator<char>(file_name)),
               istreambuf_iterator<char>());

    Local<String> source =
        String::NewFromUtf8(isolate, src.c_str(), NewStringType::kNormal)
            .ToLocalChecked();
    Local<Script> script = Script::Compile(context, source).ToLocalChecked();
    Local<Value> result = script->Run(context).ToLocalChecked();

    Local<String> handle_user_req =
        String::NewFromUtf8(isolate, "DebugUserRequest", NewStringType::kNormal)
            .ToLocalChecked();
    Local<Value> handle_user_req_val;
    if (!context->Global()
             ->Get(context, handle_user_req)
             .ToLocal(&handle_user_req_val))
      cout << "Failed to grab DebugUserRequest function " << endl;

    Local<Function> handle_user_req_fun =
        Local<Function>::Cast(handle_user_req_val);

    assert(handle_user_req_fun->IsFunction());
    debug_user_request.Reset(isolate, handle_user_req_fun);
  }

  int iter_count = atoi(argv[2]);
  std::cout << "iter_count: " << iter_count << std::endl;
  thread send_debug_user_req_thr(ProcessRequest, iter_count);
  thread terminate_thr(TerminateTask);
  sleep(3);

  send_debug_user_req_thr.join();
  terminate_thr.join();

  isolate->Dispose();
  V8::Dispose();
  V8::ShutdownPlatform();
  delete platform;
  return 0;
}
