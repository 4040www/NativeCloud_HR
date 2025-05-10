from sqlalchemy import create_engine, text
from faker import Faker
from dotenv import load_dotenv
import os

# === 連線設定 ===

# 載入 .env 檔案
load_dotenv()

# 組合 DB_URL
DB_HOST = os.getenv("DB_HOST")
DB_USER = os.getenv("DB_USER")
DB_PASSWORD = os.getenv("DB_PASSWORD")
DB_NAME = os.getenv("DB_NAME")
DB_PORT = os.getenv("DB_PORT")

DB_URL = f"postgresql://{DB_USER}:{DB_PASSWORD}@{DB_HOST}:{DB_PORT}/{DB_NAME}"
engine = create_engine(DB_URL)
fake = Faker()


# === 獲得 Employee 資料 ===
def get_employees(conn):
    
    employees = []
    employee_ids = []
    org_id = "L120"  # 假設我們只查詢組織 ID 為 L1 的員工
    # 查詢員工資料
    result = conn.execute(text(
        "SELECT employee_id, first_name, last_name, is_manager, password, email FROM employee WHERE organization_id = :org_id"
    ), {"org_id": org_id})
    employee_data = result.fetchone()
    if employee_data:
        employee_ids.append(employee_data[0])
        print(f"✅ organization {org_id}, 員工 {employee_data[0]}, is_manager={employee_data[3]}, email={employee_data[5]}, password={employee_data[4]}")
        
    return employee_ids
    

    #     result = conn.execute(text("""
    #         INSERT INTO employee (first_name, last_name, is_manager, password, email, organization_id)
    #         VALUES (:first_name, :last_name, :is_manager, :password, :email, :organization_id)
    #         RETURNING employee_id
    #     """), {
    #         "first_name": first_name,
    #         "last_name": last_name,
    #         "is_manager": is_manager,
    #         "password": password,
    #         "email": email,
    #         "organization_id": org_id
    #     })
    #     employee_id = result.fetchone()[0]
    #     employee_ids.append(employee_id)
    #     if(len(employee_ids) == 10):
    #         print(f"✅ organization {org_id}, 員工 {employee_id}, is_manager={is_manager}, email={email}, password={password}")

    # print(f"✅ organization {org_id} 已插入 {num_employees} 個 employee")
    # return employee_ids

# === 主程式：自由調用 ===
if __name__ == "__main__":
    with engine.connect() as conn:
        employee_id = get_employees(conn)
        conn.commit()
    print("查詢完成！")