import psycopg2
from dotenv import load_dotenv
import os

# 載入 .env 檔案
load_dotenv()

# 組合 DB_URL
DB_HOST = os.getenv("DB_HOST")
DB_USER = os.getenv("DB_USER")
DB_PASSWORD = os.getenv("DB_PASSWORD")
DB_NAME = os.getenv("DB_NAME")
DB_PORT = os.getenv("DB_PORT")

DB_URL = f"postgresql://{DB_USER}:{DB_PASSWORD}@{DB_HOST}:{DB_PORT}/{DB_NAME}"

create_sql = """
-- 啟用 uuid-ossp extension
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- 建立 Organization 表格
CREATE TABLE IF NOT EXISTS Organization (
    Organization_id VARCHAR(255) PRIMARY KEY,
    Name VARCHAR(255) NOT NULL
);

-- 建立 Employee 表格
CREATE TABLE IF NOT EXISTS Employee (
    Employee_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    First_name VARCHAR(255) NOT NULL,
    Last_name VARCHAR(255) NOT NULL,
    Is_manager BOOLEAN DEFAULT FALSE,
    Password VARCHAR(255) NOT NULL,
    Email VARCHAR(255) NOT NULL,
    Organization_id VARCHAR(255),
    FOREIGN KEY (Organization_id) REFERENCES Organization(Organization_id)
);

-- 建立 Access_log 表格
CREATE TABLE IF NOT EXISTS Access_log (
    Access_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    Employee_id UUID,
    Access_time TIMESTAMP NOT NULL,
    Direction VARCHAR(50),
    Gate_type VARCHAR(50),
    Gate_name VARCHAR(100),
    Access_result VARCHAR(50),
    FOREIGN KEY (Employee_id) REFERENCES Employee(Employee_id)
);
"""

def create_tables():
    try:
        conn = psycopg2.connect(DB_URL)
        conn.autocommit = True
        cursor = conn.cursor()
        cursor.execute(create_sql)
        print("Tables created successfully.")
        cursor.close()
        conn.close()
    except Exception as e:
        print("Error creating tables:", e)

if __name__ == "__main__":
    create_tables()
