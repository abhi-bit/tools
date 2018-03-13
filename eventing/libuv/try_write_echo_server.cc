#include <cstdio>
#include <cstdlib>
#include <cstring>
#include <iostream>
#include <uv.h>
#include <vector>

uv_loop_t *loop;

typedef struct {
  uv_write_t req;
  uv_buf_t buf;
} write_req_t;

void free_write_req(uv_write_t *req) {
  write_req_t *wr = (write_req_t *)req;
  free(wr->buf.base);
  free(wr);
}

void alloc_buffer(uv_handle_t *handle, size_t suggested_size, uv_buf_t *buf) {
  buf->base = (char *)malloc(suggested_size);
  buf->len = suggested_size;
}

void echo_read(uv_stream_t *client, ssize_t nread, const uv_buf_t *buf) {
  if (nread > 0) {
    int bytes_to_write = 0;
    std::vector<uv_buf_t> buffers;

    uv_buf_t buffer1 = uv_buf_init(buf->base, nread);
    bytes_to_write += buffer1.len;
    buffers.push_back(buffer1);

    uv_buf_t buffer2 = uv_buf_init(buf->base, nread);
    bytes_to_write += buffer2.len;
    buffers.push_back(buffer2);

    uv_buf_t buffer3 = uv_buf_init(buf->base, nread);
    bytes_to_write += buffer3.len;
    buffers.push_back(buffer3);

    int bytes_written = UV_EAGAIN;
    do {
      bytes_written = uv_try_write(client, buffers.data(), buffers.size());
      std::cout << "bytes written: " << bytes_written
                << " buf_base len: " << buffers.size() << " nread: " << nread
                << " bytes_to_write: " << bytes_to_write << std::endl;

      // To simulate the case where uv_try_write writes only a portion of
      // supplied buffer
      bytes_written -= (buffer3.len) + 2;

      if (bytes_written == bytes_to_write) {
          buffers.clear();
      }

      if ((bytes_written < bytes_to_write) && (bytes_written > 0)) {
        std::cout << "bytes_written: " << bytes_written
                  << " bytes_to_write: " << bytes_to_write << std::endl;
        bytes_to_write -= bytes_written;

        std::vector<uv_buf_t> temp_buffers;
        int buffer_sizes_so_far = 0;
        int index = -1;
        // Check at what std::vector index, entries were written to completely
        for (auto const &buffer : buffers) {
          std::cout << "Index: " << index << " buffer size:" << buffer.len
                    << std::endl;

          if ((buffer.len + buffer_sizes_so_far) > bytes_written) {
            std::cout << "Only entries till index: " << index
                      << " were flushed completely. Missed data from buffer: "
                      << (buffer.len + buffer_sizes_so_far) - bytes_written
                      << std::endl;

            std::string original_data(buffer.base, buffer.len);
            std::string pending_data(
                original_data,
                (buffer.len + buffer_sizes_so_far - bytes_written));
            std::cout << "original_data size: " << original_data.size()
                      << " pending_data len: " << pending_data.size()
                      << std::endl;
            uv_buf_t temp_buffer = uv_buf_init((char *)pending_data.c_str(), pending_data.size());
            temp_buffers.push_back(temp_buffer);
            ++index;
            break;
          } else {
            buffer_sizes_so_far += buffer.len;
            std::cout << "Incrementing buffer_sizes_so_far by: " << buffer.len
                      << " current val: " << buffer_sizes_so_far << std::endl;
            ++index;
          }
        }

        std::cout << "Original buffer count: " << buffers.size() << std::endl;

        for(std::vector<int>::size_type i = index+1; i != buffers.size(); i++) {
          std::cout << "Inserting index: " << i << " into temp_buffers"
                    << std::endl;
          temp_buffers.push_back(buffers[i]);
        }
        buffers.swap(temp_buffers);
        std::cout << "size of temp_vector: " << temp_buffers.size()
                  << " original buffer size: " << buffers.size() << std::endl;
      }

    } while ((bytes_written == UV_EAGAIN) || (bytes_written == 0) || (buffers.size() > 0));
    return;
  }

  if (nread < 0) {
    if (nread != UV_EOF)
      fprintf(stderr, "Read error %s\n", uv_err_name(nread));
    uv_close((uv_handle_t *)client, NULL);
  }

  free(buf->base);
}

void on_new_connection(uv_stream_t *server, int status) {
  if (status == -1) {
    // error!
    return;
  }

  uv_pipe_t *client = (uv_pipe_t *)malloc(sizeof(uv_pipe_t));
  uv_pipe_init(loop, client, 0);
  if (uv_accept(server, (uv_stream_t *)client) == 0) {
    uv_read_start((uv_stream_t *)client, alloc_buffer, echo_read);
  } else {
    uv_close((uv_handle_t *)client, NULL);
  }
}

void remove_sock(int sig) {
  uv_fs_t req;
  uv_fs_unlink(loop, &req, "echo.sock", NULL);
  exit(0);
}

int main() {
  loop = uv_default_loop();

  uv_pipe_t server;
  uv_pipe_init(loop, &server, 0);

  signal(SIGINT, remove_sock);

  int r;
  if ((r = uv_pipe_bind(&server, "echo.sock"))) {
    fprintf(stderr, "Bind error %s\n", uv_err_name(r));
    return 1;
  }
  if ((r = uv_listen((uv_stream_t *)&server, 128, on_new_connection))) {
    fprintf(stderr, "Listen error %s\n", uv_err_name(r));
    return 2;
  }
  return uv_run(loop, UV_RUN_DEFAULT);
}
