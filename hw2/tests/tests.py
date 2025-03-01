import base64
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

@pytest.fixture(scope='module')
def auth_token(user_data):
    register_url = f"{BASE_URL}/register"
    login_url = f"{BASE_URL}/login"

    response = requests.post(register_url, json=user_data)
    assert response.status_code == 201

    response = requests.post(login_url, json=user_data)
    assert response.status_code == 200

    data = response.json()
    assert 'token' in data

    return data['token']

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

def get_code_processor_payload():
    return {"tranlator": "python3", "code": "print('Hello, stdout world!')"}

def get_image_processor_payload():
    with open("static/sigma.png", "rb") as image_file:
        image_bytes = image_file.read()

    image_base64 = base64.b64encode(image_bytes).decode('utf-8')
    return {"filter": {"name": "Negative"}, "image": image_base64}

def test_create_task(auth_token):
    task_url = f"{BASE_URL}/task"
    headers = {'Authorization': f'Bearer {auth_token}'}
    
    response = requests.post(task_url, headers=headers, json=payload) 

    assert response.status_code == 201
    data = response.json()
    assert 'task_id' in data

    return data['task_id']

def test_task_status_and_result(auth_token):
    task_id = test_create_task(auth_token)
    status_url = f"{BASE_URL}/status/{task_id}"
    result_url = f"{BASE_URL}/result/{task_id}"
    headers = {'Authorization': f'Bearer {auth_token}'}

    retry = 10
    while retry >= 0:
        response = requests.get(status_url, headers=headers)
        assert response.status_code == 200
        data = response.json()
        assert 'status' in data
        
        if data['status'] == 'ready':
            break
         
        assert data['status'] == 'in_progress', f'undefined status: {data['status']}!'
        retry -= 1
        time.sleep(3)
    assert retry > 0, "task is still in progress!"

    response = requests.get(result_url, headers=headers)
    assert response.status_code == 200
    data = response.json()
    assert 'result' in data

def test_task_not_found(auth_token):
    invalid_task_id = str(uuid.uuid4())
    status_url = f"{BASE_URL}/status/{invalid_task_id}"
    result_url = f"{BASE_URL}/result/{invalid_task_id}"
    headers = {'Authorization': f'Bearer {auth_token}'}

    response = requests.get(status_url, headers=headers)
    assert response.status_code == 404

    response = requests.get(result_url, headers=headers)
    assert response.status_code == 404

def test_unauthorized_access():
    invalid_task_id = str(uuid.uuid4())
    status_url = f"{BASE_URL}/status/{invalid_task_id}"
    result_url = f"{BASE_URL}/result/{invalid_task_id}"
    task_url = f"{BASE_URL}/task"

    response = requests.post(task_url)
    assert response.status_code == 401

    response = requests.get(status_url)
    assert response.status_code == 401

    response = requests.get(result_url)
    assert response.status_code == 401
