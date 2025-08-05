#!/usr/bin/env python3
import os
import sys
import time
import subprocess
import threading
import signal
import json
import requests
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

class CryptoServerTestRunner:
    def __init__(self):
        self.server_process = None
        self.server_port = 8080
        self.server_url = f"http://localhost:{self.server_port}"
        self.test_results = []
        self.compilation_failed = False
        self.failure_reason = ""
        self.auth_token = None
        self.test_username = "testuser_" + str(int(time.time()))
        self.test_password = "testpass123"

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
        required_scripts = ['compile.sh', 'execute.sh']
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
        self.info("Компиляция кода...")
        result = subprocess.run(['./compile.sh'], capture_output=True, text=True)
        if result.returncode != 0:
            self.compilation_failed = True
            self.failure_reason = f"Ошибка компиляции: {result.stderr.strip() or result.stdout.strip() or 'Неизвестная ошибка'}"
            self.error(self.failure_reason)
            return False

        self.success("Компиляция завершена успешно")
        return True

    def start_server(self):
        try:
            self.info("Запуск crypto сервера...")

            self.server_process = subprocess.Popen(
                ['./execute.sh'],
                stdout=subprocess.PIPE,
                stderr=subprocess.PIPE,
                preexec_fn=os.setsid if os.name != 'nt' else None
            )

            time.sleep(3)

            if self.server_process.poll() is not None:
                stdout, stderr = self.server_process.communicate()
                self.failure_reason = f"Сервер завершился с ошибкой: {stderr.decode().strip() or stdout.decode().strip() or 'Неизвестная ошибка'}"
                self.error(self.failure_reason)
                return False

            return self.check_server_responding()

        except FileNotFoundError:
            self.failure_reason = "Не удалось запустить execute.sh"
            self.error(self.failure_reason)
            return False
        except Exception as e:
            self.failure_reason = f"Ошибка запуска сервера: {e}"
            self.error(self.failure_reason)
            return False

    def check_server_responding(self):
        for _ in range(10):
            try:
                requests.get(f"{self.server_url}/", timeout=2)
                self.success(f"Сервер отвечает на порту {self.server_port}")
                return True
            except requests.exceptions.RequestException:
                time.sleep(1)

        try:
            requests.get(f"{self.server_url}/crypto", timeout=2)
            self.success(f"Сервер отвечает на порту {self.server_port}")
            return True
        except requests.exceptions.RequestException:
            pass

        self.failure_reason = f"Сервер не отвечает на порту {self.server_port}"
        self.error(self.failure_reason)
        return False

    def test_user_registration(self):
        try:
            self.info("Тестирование регистрации пользователя...")

            registration_data = {
                "username": self.test_username,
                "password": self.test_password
            }

            response = requests.post(
                f"{self.server_url}/auth/register",
                json=registration_data,
                timeout=5
            )

            if response.status_code not in [200, 201]:
                self.error(f"Неверный статус код при регистрации: {response.status_code}")
                return False

            try:
                data = response.json()
            except:
                self.error("Ответ не является валидным JSON")
                return False

            if "error" in data:
                self.error(f"Регистрация не удалась: {data['error']}")
                return False

            if not data.get("token"):
                self.error("Не получен токен при регистрации")
                return False

            self.auth_token = data["token"]
            self.success("Регистрация пользователя прошла успешно")

            response = requests.post(
                f"{self.server_url}/auth/register",
                json=registration_data,
                timeout=5
            )

            if response.status_code != 409:
                self.warning(f"Ожидался статус 409 при повторной регистрации, получен: {response.status_code}")

            return True

        except requests.exceptions.RequestException as e:
            self.error(f"Ошибка при тестировании регистрации: {e}")
            return False
        except Exception as e:
            self.error(f"Неожиданная ошибка при регистрации: {e}")
            return False

    def test_user_login(self):
        try:
            self.info("Тестирование входа пользователя...")

            login_data = {
                "username": self.test_username,
                "password": self.test_password
            }

            response = requests.post(
                f"{self.server_url}/auth/login",
                json=login_data,
                timeout=5
            )

            if response.status_code != 200:
                self.error(f"Неверный статус код при входе: {response.status_code}")
                return False

            try:
                data = response.json()
            except:
                self.error("Ответ не является валидным JSON")
                return False

            if "error" in data:
                self.error(f"Вход не удался: {data['error']}")
                return False

            if not data.get("token"):
                self.error("Не получен токен при входе")
                return False

            self.auth_token = data["token"]
            self.success("Вход пользователя прошел успешно")

            wrong_login_data = {
                "username": self.test_username,
                "password": "wrongpassword"
            }

            response = requests.post(
                f"{self.server_url}/auth/login",
                json=wrong_login_data,
                timeout=5
            )

            if response.status_code not in [401, 400]:
                self.warning(f"Ожидался статус 401/400 при неверном пароле, получен: {response.status_code}")

            return True

        except requests.exceptions.RequestException as e:
            self.error(f"Ошибка при тестировании входа: {e}")
            return False
        except Exception as e:
            self.error(f"Неожиданная ошибка при входе: {e}")
            return False

    def test_add_crypto(self):
        try:
            self.info("Тестирование добавления криптовалюты...")

            if not self.auth_token:
                self.error("Нет токена для аутентификации")
                return False

            headers = {"Authorization": f"Bearer {self.auth_token}"}
            crypto_data = {"symbol": "BTC"}

            response = requests.post(
                f"{self.server_url}/crypto",
                json=crypto_data,
                headers=headers,
                timeout=10
            )

            if response.status_code not in [200, 201]:
                self.error(f"Неверный статус код при добавлении криптовалюты: {response.status_code}")
                self.error(f"Ответ: {response.text}")
                return False

            try:
                data = response.json()
            except:
                self.error("Ответ не является валидным JSON")
                return False

            if "error" in data:
                self.error(f"Добавление криптовалюты не удалось: {data['error']}")
                return False

            crypto_info = data.get("crypto")
            if not crypto_info:
                self.error("Не получена информация о криптовалюте")
                return False

            required_fields = ["symbol", "name", "current_price", "last_updated"]
            for field in required_fields:
                if field not in crypto_info:
                    self.error(f"Отсутствует поле '{field}' в информации о криптовалюте")
                    return False

            self.success(f"Криптовалюта {crypto_info['symbol']} добавлена успешно")
            self.info(f"Цена: ${crypto_info['current_price']}")

            response = requests.post(
                f"{self.server_url}/crypto",
                json=crypto_data,
                headers=headers,
                timeout=5
            )

            if response.status_code != 409:
                self.warning(f"Ожидался статус 409 при повторном добавлении, получен: {response.status_code}")

            return True

        except requests.exceptions.RequestException as e:
            self.error(f"Ошибка при добавлении криптовалюты: {e}")
            return False
        except Exception as e:
            self.error(f"Неожиданная ошибка при добавлении криптовалюты: {e}")
            return False

    def test_get_crypto_list(self):
        try:
            self.info("Тестирование получения списка криптовалют...")

            if not self.auth_token:
                self.error("Нет токена для аутентификации")
                return False

            headers = {"Authorization": f"Bearer {self.auth_token}"}

            response = requests.get(
                f"{self.server_url}/crypto",
                headers=headers,
                timeout=5
            )

            if response.status_code != 200:
                self.error(f"Неверный статус код при получении списка: {response.status_code}")
                return False

            try:
                data = response.json()
            except:
                self.error("Ответ не является валидным JSON")
                return False

            if "cryptos" not in data:
                self.error("Отсутствует поле 'cryptos' в ответе")
                return False

            cryptos = data["cryptos"]
            if not isinstance(cryptos, list):
                self.error("Поле 'cryptos' должно быть массивом")
                return False

            if len(cryptos) == 0:
                self.warning("Список криптовалют пуст")
            else:
                self.success(f"Получен список из {len(cryptos)} криптовалют")

                first_crypto = cryptos[0]
                required_fields = ["symbol", "name", "current_price", "last_updated"]
                for field in required_fields:
                    if field not in first_crypto:
                        self.error(f"Отсутствует поле '{field}' в информации о криптовалюте")
                        return False

            return True

        except requests.exceptions.RequestException as e:
            self.error(f"Ошибка при получении списка криптовалют: {e}")
            return False
        except Exception as e:
            self.error(f"Неожиданная ошибка при получении списка: {e}")
            return False

    def test_get_specific_crypto(self):
        try:
            self.info("Тестирование получения конкретной криптовалюты...")

            if not self.auth_token:
                self.error("Нет токена для аутентификации")
                return False

            headers = {"Authorization": f"Bearer {self.auth_token}"}

            response = requests.get(
                f"{self.server_url}/crypto/BTC",
                headers=headers,
                timeout=5
            )

            if response.status_code != 200:
                self.error(f"Неверный статус код при получении BTC: {response.status_code}")
                return False

            try:
                data = response.json()
            except:
                self.error("Ответ не является валидным JSON")
                return False

            required_fields = ["symbol", "name", "current_price", "last_updated"]
            for field in required_fields:
                if field not in data:
                    self.error(f"Отсутствует поле '{field}' в информации о криптовалюте")
                    return False

            self.success(f"Получена информация о {data['symbol']}: ${data['current_price']}")

            response = requests.get(
                f"{self.server_url}/crypto/NONEXISTENT",
                headers=headers,
                timeout=5
            )

            if response.status_code != 404:
                self.warning(f"Ожидался статус 404 для несуществующей криптовалюты, получен: {response.status_code}")

            return True

        except requests.exceptions.RequestException as e:
            self.error(f"Ошибка при получении конкретной криптовалюты: {e}")
            return False
        except Exception as e:
            self.error(f"Неожиданная ошибка при получении криптовалюты: {e}")
            return False

    def test_refresh_crypto_price(self):
        try:
            self.info("Тестирование обновления цены криптовалюты...")

            if not self.auth_token:
                self.error("Нет токена для аутентификации")
                return False

            headers = {"Authorization": f"Bearer {self.auth_token}"}

            response = requests.put(
                f"{self.server_url}/crypto/BTC/refresh",
                headers=headers,
                timeout=10
            )

            if response.status_code != 200:
                self.error(f"Неверный статус код при обновлении цены: {response.status_code}")
                return False

            try:
                data = response.json()
            except:
                self.error("Ответ не является валидным JSON")
                return False

            if "error" in data:
                self.error(f"Обновление цены не удалось: {data['error']}")
                return False

            crypto_info = data.get("crypto")
            if not crypto_info:
                self.error("Не получена обновленная информация о криптовалюте")
                return False

            self.success(f"Цена {crypto_info['symbol']} обновлена: ${crypto_info['current_price']}")
            return True

        except requests.exceptions.RequestException as e:
            self.error(f"Ошибка при обновлении цены: {e}")
            return False
        except Exception as e:
            self.error(f"Неожиданная ошибка при обновлении цены: {e}")
            return False

    def test_crypto_price_history(self):
        try:
            self.info("Тестирование истории цен криптовалюты...")

            if not self.auth_token:
                self.error("Нет токена для аутентификации")
                return False

            headers = {"Authorization": f"Bearer {self.auth_token}"}

            response = requests.get(
                f"{self.server_url}/crypto/BTC/history",
                headers=headers,
                timeout=5
            )

            if response.status_code != 200:
                self.error(f"Неверный статус код при получении истории: {response.status_code}")
                return False

            try:
                data = response.json()
            except:
                self.error("Ответ не является валидным JSON")
                return False

            if "symbol" not in data or "history" not in data:
                self.error("Отсутствуют обязательные поля в ответе истории")
                return False

            history = data["history"]
            if not isinstance(history, list):
                self.error("История должна быть массивом")
                return False

            if len(history) > 0:
                first_entry = history[0]
                if "price" not in first_entry or "timestamp" not in first_entry:
                    self.error("Неверная структура записи в истории")
                    return False

            self.success(f"Получена история для BTC: {len(history)} записей")
            return True

        except requests.exceptions.RequestException as e:
            self.error(f"Ошибка при получении истории: {e}")
            return False
        except Exception as e:
            self.error(f"Неожиданная ошибка при получении истории: {e}")
            return False

    def test_crypto_stats(self):
        try:
            self.info("Тестирование статистики криптовалюты...")

            if not self.auth_token:
                self.error("Нет токена для аутентификации")
                return False

            headers = {"Authorization": f"Bearer {self.auth_token}"}

            response = requests.get(
                f"{self.server_url}/crypto/BTC/stats",
                headers=headers,
                timeout=5
            )

            if response.status_code != 200:
                self.error(f"Неверный статус код при получении статистики: {response.status_code}")
                return False

            try:
                data = response.json()
            except:
                self.error("Ответ не является валидным JSON")
                return False

            if "symbol" not in data or "current_price" not in data or "stats" not in data:
                self.error("Отсутствуют обязательные поля в ответе статистики")
                return False

            stats = data["stats"]
            if isinstance(stats, dict):
                expected_fields = ["min_price", "max_price", "avg_price", "records_count"]
                for field in expected_fields:
                    if field not in stats:
                        self.error(f"Отсутствует поле '{field}' в статистике")
                        return False
                self.success(f"Получена статистика для BTC: {stats['records_count']} записей")
            else:
                self.success("Получено сообщение о недостатке данных для статистики")

            return True

        except requests.exceptions.RequestException as e:
            self.error(f"Ошибка при получении статистики: {e}")
            return False
        except Exception as e:
            self.error(f"Неожиданная ошибка при получении статистики: {e}")
            return False

    def test_delete_crypto(self):
        try:
            self.info("Тестирование удаления криптовалюты...")

            if not self.auth_token:
                self.error("Нет токена для аутентификации")
                return False

            headers = {"Authorization": f"Bearer {self.auth_token}"}

            crypto_data = {"symbol": "ETH"}
            response = requests.post(
                f"{self.server_url}/crypto",
                json=crypto_data,
                headers=headers,
                timeout=10
            )

            response = requests.delete(
                f"{self.server_url}/crypto/ETH",
                headers=headers,
                timeout=5
            )

            if response.status_code != 200:
                self.error(f"Неверный статус код при удалении: {response.status_code}")
                return False

            try:
                data = response.json()
            except:
                self.error("Ответ не является валидным JSON")
                return False

            if "error" in data:
                self.error(f"Удаление не удалось: {data['error']}")
                return False

            self.success("Криптовалюта удалена успешно")

            response = requests.get(
                f"{self.server_url}/crypto/ETH",
                headers=headers,
                timeout=5
            )

            if response.status_code != 404:
                self.warning(f"Ожидался статус 404 для удаленной криптовалюты, получен: {response.status_code}")

            return True

        except requests.exceptions.RequestException as e:
            self.error(f"Ошибка при удалении криптовалюты: {e}")
            return False
        except Exception as e:
            self.error(f"Неожиданная ошибка при удалении: {e}")
            return False

    def test_authentication_required(self):
        try:
            self.info("Тестирование требования аутентификации...")

            endpoints_to_test = [
                ("GET", "/crypto"),
                ("POST", "/crypto"),
                ("GET", "/crypto/BTC"),
                ("PUT", "/crypto/BTC/refresh"),
                ("DELETE", "/crypto/BTC")
            ]

            schedule_enabled = os.environ.get('SCHEDULE') == '1'
            if schedule_enabled:
                endpoints_to_test.extend([
                    ("GET", "/schedule"),
                    ("PUT", "/schedule"),
                    ("POST", "/schedule/trigger")
                ])

            for method, endpoint in endpoints_to_test:
                if method == "GET":
                    response = requests.get(f"{self.server_url}{endpoint}", timeout=5)
                elif method == "POST":
                    if endpoint == "/crypto":
                        response = requests.post(f"{self.server_url}{endpoint}", json={"symbol": "BTC"}, timeout=5)
                    else:  # /schedule/trigger
                        response = requests.post(f"{self.server_url}{endpoint}", timeout=5)
                elif method == "PUT":
                    if endpoint == "/schedule":
                        response = requests.put(f"{self.server_url}{endpoint}", json={"enabled": True}, timeout=5)
                    else:
                        response = requests.put(f"{self.server_url}{endpoint}", timeout=5)
                elif method == "DELETE":
                    response = requests.delete(f"{self.server_url}{endpoint}", timeout=5)

                if response.status_code != 401:
                    self.warning(f"Ожидался статус 401 для {method} {endpoint} без токена, получен: {response.status_code}")

            self.success("Проверка требования аутентификации завершена")
            return True

        except requests.exceptions.RequestException as e:
            self.error(f"Ошибка при тестировании аутентификации: {e}")
            return False
        except Exception as e:
            self.error(f"Неожиданная ошибка при тестировании аутентификации: {e}")
            return False

    def test_schedule_get(self):
        try:
            self.info("Тестирование получения настроек расписания...")

            if not self.auth_token:
                self.error("Нет токена для аутентификации")
                return False

            headers = {"Authorization": f"Bearer {self.auth_token}"}

            response = requests.get(
                f"{self.server_url}/schedule",
                headers=headers,
                timeout=5
            )

            if response.status_code != 200:
                self.error(f"Неверный статус код при получении расписания: {response.status_code}")
                return False

            try:
                data = response.json()
            except:
                self.error("Ответ не является валидным JSON")
                return False

            if "error" in data:
                self.error(f"Получение расписания не удалось: {data['error']}")
                return False

            required_fields = ["enabled", "interval_seconds"]
            for field in required_fields:
                if field not in data:
                    self.error(f"Отсутствует поле '{field}' в настройках расписания")
                    return False

            self.success(f"Получены настройки расписания: enabled={data['enabled']}, interval={data['interval_seconds']}s")
            return True

        except requests.exceptions.RequestException as e:
            self.error(f"Ошибка при получении расписания: {e}")
            return False
        except Exception as e:
            self.error(f"Неожиданная ошибка при получении расписания: {e}")
            return False

    def test_schedule_update(self):
        try:
            self.info("Тестирование изменения настроек расписания...")

            if not self.auth_token:
                self.error("Нет токена для аутентификации")
                return False

            headers = {"Authorization": f"Bearer {self.auth_token}", "Content-Type": "application/json"}

            schedule_data = {"enabled": True, "interval_seconds": 60}
            response = requests.put(
                f"{self.server_url}/schedule",
                json=schedule_data,
                headers=headers,
                timeout=5
            )

            if response.status_code != 200:
                self.error(f"Неверный статус код при изменении расписания: {response.status_code}")
                return False

            try:
                data = response.json()
            except:
                self.error("Ответ не является валидным JSON")
                return False

            if "error" in data:
                self.error(f"Изменение расписания не удалось: {data['error']}")
                return False

            if data.get("interval_seconds") != 60 or data.get("enabled") != True:
                self.error(f"Настройки не изменились корректно: {data}")
                return False

            invalid_schedule = {"interval_seconds": 5}
            response = requests.put(
                f"{self.server_url}/schedule",
                json=invalid_schedule,
                headers=headers,
                timeout=5
            )

            if response.status_code != 400:
                self.warning(f"Ожидался статус 400 для некорректного интервала, получен: {response.status_code}")

            self.success("Настройки расписания изменены успешно")
            return True

        except requests.exceptions.RequestException as e:
            self.error(f"Ошибка при изменении расписания: {e}")
            return False
        except Exception as e:
            self.error(f"Неожиданная ошибка при изменении расписания: {e}")
            return False

    def test_schedule_trigger(self):
        try:
            self.info("Тестирование принудительного обновления цен...")

            if not self.auth_token:
                self.error("Нет токена для аутентификации")
                return False

            headers = {"Authorization": f"Bearer {self.auth_token}"}

            crypto_data = {"symbol": "BTC"}
            requests.post(
                f"{self.server_url}/crypto",
                json=crypto_data,
                headers=headers,
                timeout=10
            )

            response = requests.post(
                f"{self.server_url}/schedule/trigger",
                headers=headers,
                timeout=10
            )

            if response.status_code != 200:
                self.error(f"Неверный статус код при принудительном обновлении: {response.status_code}")
                return False

            try:
                data = response.json()
            except:
                self.error("Ответ не является валидным JSON")
                return False

            if "error" in data:
                self.error(f"Принудительное обновление не удалось: {data['error']}")
                return False

            required_fields = ["updated_count", "timestamp"]
            for field in required_fields:
                if field not in data:
                    self.error(f"Отсутствует поле '{field}' в ответе принудительного обновления")
                    return False

            updated_count = data["updated_count"]
            if updated_count < 0:
                self.error(f"Некорректное количество обновленных криптовалют: {updated_count}")
                return False

            self.success(f"Принудительное обновление выполнено для {updated_count} криптовалют")
            return True

        except requests.exceptions.RequestException as e:
            self.error(f"Ошибка при принудительном обновлении: {e}")
            return False
        except Exception as e:
            self.error(f"Неожиданная ошибка при принудительном обновлении: {e}")
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
        self.log("🧪 Начало тестирования домашнего задания №2", Fore.CYAN)

        if not self.check_scripts_exist():
            return False

        if not self.compile_code():
            return False

        try:
            if not self.start_server():
                return False

            tests = [
                ("Регистрация пользователя", self.test_user_registration),
                ("Вход пользователя", self.test_user_login),
                ("Добавление криптовалюты", self.test_add_crypto),
                ("Получение списка криптовалют", self.test_get_crypto_list),
                ("Получение конкретной криптовалюты", self.test_get_specific_crypto),
                ("Обновление цены криптовалюты", self.test_refresh_crypto_price),
                ("История цен криптовалюты", self.test_crypto_price_history),
                ("Статистика криптовалюты", self.test_crypto_stats),
                ("Удаление криптовалюты", self.test_delete_crypto),
                ("Требование аутентификации", self.test_authentication_required)
            ]

            if os.environ.get('SCHEDULE') == '1':
                self.log("🚨 Включены дополнительные тесты : Расписание автоматического обновления", Fore.YELLOW)
                schedule_tests = [
                    ("Получение настроек расписания", self.test_schedule_get),
                    ("Изменение настроек расписания", self.test_schedule_update),
                    ("Принудительное обновление цен", self.test_schedule_trigger),
                ]
                # append test before deleting crypto
                tests = tests[:-2] + schedule_tests + tests[-2:]
            else:
                self.log("ℹ️  Дополнительные тесты отключены (используйте SCHEDULE=1 для включения)", Fore.CYAN)

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

        schedule_enabled = os.environ.get('SCHEDULE') == '1'
        if schedule_enabled:
            self.log(f"📋 Режим: Полные тесты (основные + дополнительные)", Fore.CYAN)
        else:
            self.log(f"📋 Режим: Основные тесты (используйте SCHEDULE=1 для полного тестирования)", Fore.CYAN)

        if passed == total:
            self.success(f"Все тесты пройдены: {passed}/{total}")
            if schedule_enabled:
                self.log("🎉 Поздравляем! Домашнее задание выполнено корректно с дополнительными функциями!", Fore.GREEN)
            else:
                self.log("🎉 Поздравляем! Домашнее задание выполнено корректно!", Fore.GREEN)
        else:
            self.error(f"Тесты пройдены: {passed}/{total}")
            self.log("❌ Есть ошибки, которые нужно исправить", Fore.RED)

def main():
    os.chdir(Path(__file__).parent.parent)

    runner = CryptoServerTestRunner()

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