import pytest
import requests
import uuid
import time

BASE_URL = "http://127.0.0.1:8000"

@pytest.fixture(scope='module')
def user_data():
    username = f'user_{uuid.uuid4()}'
    password = 'password228'
    return {'username': username, 'password': password}

def test_register_user(user_data):
    register_url = f"{BASE_URL}/register"
    response = requests.post(register_url, json=user_data)
    assert response.status_code == 201

def test_login_user(user_data):
    login_url = f"{BASE_URL}/login"
    response = requests.post(login_url, json=user_data)
    assert response.status_code == 200
    data = response.json()
    assert 'token' in data

def test_create_task():
    task_url = f"{BASE_URL}/task"
    response = requests.post(task_url)

    assert response.status_code == 201
    data = response.json()
    assert 'task_id' in data

    return data['task_id']

def test_task_status_and_result():
    task_id = test_create_task()
    status_url = f"{BASE_URL}/status/{task_id}"
    result_url = f"{BASE_URL}/result/{task_id}"

    while True:
        response = requests.get(status_url)
        assert response.status_code == 200
        data = response.json()
        assert 'status' in data
        
        if data['status'] == 'ready':
            break
         
        assert data['status'] == 'in_progress', f'undefined status: {data['status']}!'
        time.sleep(1)

    response = requests.get(result_url)
    assert response.status_code == 200
    data = response.json()
    assert 'result' in data

def test_task_not_found():
    invalid_task_id = str(uuid.uuid4())
    status_url = f"{BASE_URL}/status/{invalid_task_id}"
    result_url = f"{BASE_URL}/result/{invalid_task_id}"

    response = requests.get(status_url)
    assert response.status_code == 404

    response = requests.get(result_url)
    assert response.status_code == 404
