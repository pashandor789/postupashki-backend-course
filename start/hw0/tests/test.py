#!/usr/bin/env python3
import os
import sys
import time
import socket
import subprocess
import threading
import signal
from pathlib import Path

try:
    import psutil
    PSUTIL_AVAILABLE = True
except ImportError:
    PSUTIL_AVAILABLE = False

try:
    from colorama import init, Fore, Style
    init()
    COLORS_AVAILABLE = True
except ImportError:
    COLORS_AVAILABLE = False

class TestRunner:
    def __init__(self):
        self.server_process = None
        self.server_port = 8080
        self.test_results = []
        self.compilation_failed = False
        self.failure_reason = ""

    def log(self, message, color=None):
        if COLORS_AVAILABLE and color:
            print(f"{color}{message}{Style.RESET_ALL}")
        else:
            print(message)

    def success(self, message):
        self.log(f"✅ {message}", Fore.GREEN)

    def error(self, message):
        self.log(f"❌ {message}", Fore.RED)

    def info(self, message):
        self.log(f"ℹ️  {message}", Fore.BLUE)

    def warning(self, message):
        self.log(f"⚠️  {message}", Fore.YELLOW)

    def check_scripts_exist(self):
        required_scripts = ['compile_server.sh', 'execute_server.sh', 'compile_client.sh', 'execute_client.sh']
        missing_scripts = []

        for script in required_scripts:
            if not os.path.exists(script):
                missing_scripts.append(script)

        if missing_scripts:
            self.failure_reason = f"Отсутствуют скрипты: {', '.join(missing_scripts)}"
            self.error(self.failure_reason)
            return False

        return True

    def compile_code(self):
        self.info("Компиляция серверного кода...")
        result = subprocess.run(['./compile_server.sh'], capture_output=True, text=True)
        if result.returncode != 0:
            self.compilation_failed = True
            self.failure_reason = f"Ошибка компиляции сервера: {result.stderr.strip() or result.stdout.strip() or 'Неизвестная ошибка'}"
            self.error(self.failure_reason)
            return False

        self.info("Компиляция клиентского кода...")
        result = subprocess.run(['./compile_client.sh'], capture_output=True, text=True)
        if result.returncode != 0:
            self.compilation_failed = True
            self.failure_reason = f"Ошибка компиляции клиента: {result.stderr.strip() or result.stdout.strip() or 'Неизвестная ошибка'}"
            self.error(self.failure_reason)
            return False

        self.success("Компиляция завершена успешно")
        return True

    def start_server(self):
        try:
            self.info("Запуск сервера...")

            self.server_process = subprocess.Popen(
                ['./execute_server.sh'],
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                preexec_fn=os.setsid if os.name != 'nt' else None
            )

            time.sleep(2)

            if self.server_process.poll() is not None:
                stdout, stderr = self.server_process.communicate()
                self.failure_reason = f"Сервер завершился с ошибкой: {stderr.decode().strip() or stdout.decode().strip() or 'Неизвестная ошибка'}"
                self.error(self.failure_reason)
                return False

            return self.check_port_open()

        except FileNotFoundError:
            self.failure_reason = "Не удалось запустить execute_server.sh"
            self.error(self.failure_reason)
            return False
        except Exception as e:
            self.failure_reason = f"Ошибка запуска сервера: {e}"
            self.error(self.failure_reason)
            return False

    def check_port_open(self):
        for attempt in range(5):
            try:
                with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
                    s.settimeout(1)
                    result = s.connect_ex(('localhost', self.server_port))
                    if result == 0:
                        self.success(f"Сервер слушает порт {self.server_port}")
                        return True
            except Exception:
                pass
            time.sleep(1)

        self.failure_reason = f"Сервер не отвечает на порту {self.server_port}"
        self.error(self.failure_reason)
        return False

    def test_server_response(self):
        try:
            with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
                s.settimeout(5)
                s.connect(('localhost', self.server_port))

                data = s.recv(1024)
                response = data.decode('utf-8')

                if response == "OK\n":
                    self.success("Сервер возвращает корректный ответ 'OK\\n'")
                    return True
                else:
                    self.error(f"Неверный ответ сервера. Ожидался 'OK\\n', получен: {repr(response)}")
                    return False

        except socket.timeout:
            self.error("Таймаут при подключении к серверу")
            return False
        except ConnectionRefusedError:
            self.error("Сервер отклонил соединение")
            return False
        except Exception as e:
            self.error(f"Ошибка при тестировании сервера: {e}")
            return False

    def test_multiple_connections(self):
        self.info("Тестирование множественных подключений...")

        success_count = 0
        total_tests = 3

        for i in range(total_tests):
            try:
                with socket.socket(socket.AF_INET, socket.SOCK_STREAM) as s:
                    s.settimeout(3)
                    s.connect(('localhost', self.server_port))
                    data = s.recv(1024)
                    response = data.decode('utf-8')

                    if response == "OK\n":
                        success_count += 1

                time.sleep(0.5)

            except Exception as e:
                self.warning(f"Ошибка при подключении #{i+1}: {e}")

        if success_count == total_tests:
            self.success(f"Все {total_tests} подключения успешны")
            return True
        else:
            self.error(f"Успешно только {success_count} из {total_tests} подключений")
            return False

    def test_client(self):
        try:
            self.info("Запуск клиента...")

            result = subprocess.run(['./execute_client.sh'], capture_output=True, text=True, timeout=10)

            if result.returncode == 0:
                self.success("Клиент выполнился успешно")
                self.info(f"Вывод клиента: {result.stdout.strip()}")
                return True
            else:
                self.error(f"Клиент завершился с ошибкой: {result.stderr}")
                return False

        except subprocess.TimeoutExpired:
            self.error("Таймаут выполнения клиента")
            return False
        except FileNotFoundError:
            self.error("Не удалось запустить execute_client.sh")
            return False
        except Exception as e:
            self.error(f"Ошибка при запуске клиента: {e}")
            return False

    def stop_server(self):
        if self.server_process:
            try:
                if os.name != 'nt':
                    os.killpg(os.getpgid(self.server_process.pid), signal.SIGTERM)
                else:
                    self.server_process.terminate()

                try:
                    self.server_process.wait(timeout=5)
                except subprocess.TimeoutExpired:
                    if os.name != 'nt':
                        os.killpg(os.getpgid(self.server_process.pid), signal.SIGKILL)
                    else:
                        self.server_process.kill()

                self.info("Сервер остановлен")

            except Exception as e:
                self.warning(f"Ошибка при остановке сервера: {e}")

    def run_tests(self):
        self.log("🧪 Начало тестирования домашнего задания №0", Fore.CYAN)

        if not self.check_scripts_exist():
            return False

        if not self.compile_code():
            return False

        try:
            if not self.start_server():
                return False

            tests = [
                ("Тест ответа сервера", self.test_server_response),
                ("Тест множественных подключений", self.test_multiple_connections),
                ("Тест клиента", self.test_client)
            ]

            all_passed = True
            for test_name, test_func in tests:
                self.info(f"Выполнение: {test_name}")
                if test_func():
                    self.test_results.append((test_name, True))
                else:
                    self.test_results.append((test_name, False))
                    all_passed = False
                print()

            return all_passed

        finally:
            self.stop_server()

    def print_summary(self):
        print("=" * 50)
        self.log("📊 ИТОГОВЫЙ ОТЧЁТ", Fore.CYAN)
        print("=" * 50)

        if self.compilation_failed or self.failure_reason:
            if self.failure_reason:
                self.error(f"ПРОВАЛ: {self.failure_reason}")
            else:
                self.error("ПРОВАЛ: Критическая ошибка")
            print("-" * 50)
            self.log("❌ Домашнее задание НЕ выполнено корректно", Fore.RED)
            return

        passed = sum(1 for _, result in self.test_results if result)
        total = len(self.test_results)

        if total == 0:
            self.error("ПРОВАЛ: Тесты не были запущены")
            print("-" * 50)
            self.log("❌ Домашнее задание НЕ выполнено корректно", Fore.RED)
            return

        for test_name, result in self.test_results:
            if result:
                self.success(f"{test_name}")
            else:
                self.error(f"{test_name}")

        print("-" * 50)
        if passed == total:
            self.success(f"Все тесты пройдены: {passed}/{total}")
            self.log("🎉 Поздравляем! Домашнее задание выполнено корректно!", Fore.GREEN)
        else:
            self.error(f"Тесты пройдены: {passed}/{total}")
            self.log("❌ Есть ошибки, которые нужно исправить", Fore.RED)

def main():
    os.chdir(Path(__file__).parent.parent)

    runner = TestRunner()

    try:
        success = runner.run_tests()
        runner.print_summary()
        sys.exit(0 if success else 1)

    except KeyboardInterrupt:
        runner.log("\n⏹️  Тестирование прервано пользователем", Fore.YELLOW)
        runner.stop_server()
        sys.exit(1)
    except Exception as e:
        runner.error(f"Критическая ошибка: {e}")
        runner.stop_server()
        sys.exit(1)

if __name__ == "__main__":
    main()