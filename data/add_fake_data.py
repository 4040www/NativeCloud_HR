from sqlalchemy import create_engine, text
from faker import Faker
import random
from datetime import datetime, timedelta
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

# === 建立 Organization ===
def insert_organizations(conn, num_orgs=3):
    org_ids = []
    for _ in range(num_orgs):
        name = fake.company()
        result = conn.execute(text(
            "INSERT INTO organization (name) VALUES (:name) RETURNING organization_id"
        ), {"name": name})
        org_id = result.fetchone()[0]
        org_ids.append(org_id)
    print(f"✅ 已插入 {num_orgs} 個 organization")
    return org_ids

def insert_fixed_organizations(conn):
    organizations = [
        {"name": "CEO", "organization_id": "L1"},
        {"name": "COO", "organization_id": "L10"},
        {"name": "HR Manager", "organization_id": "L100"},
        {"name": "Operations Manager", "organization_id": "L101"},
        {"name": "CFO", "organization_id": "L11"},
        {"name": "Accounting Team", "organization_id": "L110"},
        {"name": "Finance Team", "organization_id": "L111"},
        {"name": "CTO", "organization_id": "L12"},
        {"name": "Engineering Team", "organization_id": "L120"},
        {"name": "IT Support", "organization_id": "L121"},
    ]

    for org in organizations:
        conn.execute(text("""
            INSERT INTO organization (organization_id, name)
            VALUES (:organization_id, :name)
        """), org)
    
    print(f"✅ 已插入固定的 10 個 Organization！")


# === 建立 Employee ===
def insert_employees(conn, org_id, num_employees=5, create_manager=True):
    """
    為單一 org_id 插入員工資料（包含 Email 欄位）

    Args:
        conn: 資料庫連線
        org_id: 要插入的 organization_id
        num_employees: 插入的員工數量
        create_manager: 是否在第一位員工設定 is_manager=True
    """
    employee_ids = []
    manager_created = not create_manager  # 如果不需要 manager，預設已經有 manager

    for _ in range(num_employees):
        is_manager = False
        if not manager_created:
            is_manager = True
            manager_created = True

        first_name = fake.first_name()
        last_name = fake.last_name()
        email = fake.email()
        password = fake.password(length=12)

        result = conn.execute(text("""
            INSERT INTO employee (first_name, last_name, is_manager, password, email, organization_id)
            VALUES (:first_name, :last_name, :is_manager, :password, :email, :organization_id)
            RETURNING employee_id
        """), {
            "first_name": first_name,
            "last_name": last_name,
            "is_manager": is_manager,
            "password": password,
            "email": email,
            "organization_id": org_id
        })
        employee_id = result.fetchone()[0]
        employee_ids.append(employee_id)
        if(len(employee_ids) == 10):
            print(f"✅ organization {org_id}, 員工 {employee_id}, is_manager={is_manager}, email={email}, password={password}")

    print(f"✅ organization {org_id} 已插入 {num_employees} 個 employee")
    return employee_ids



# === 建立 Access_log ===
def insert_access_logs(conn, employee_ids, num_days=30):
    """
    為每個員工插入打卡紀錄 (Access_log)

    Args:
        conn: 資料庫連線
        employee_ids: 要插入的 employee_id 列表
        num_days: 每個人要插入幾天的打卡紀錄
    """
    gate_names = ["AZ_door_1", "AZ_door_2", "AZ_door_3", "AZ_door_4", "AZ_door_5", "AZ_door_6", "AZ_door_7", "AZ_door_8", "AZ_door_9", "AZ_door_10"]

    for emp_id in employee_ids:
        for i in range(num_days):
            date = datetime.today().date() - timedelta(days=i)

            # 上班打卡 (IN)
            check_in_time = datetime.combine(date, datetime.min.time()) + timedelta(hours=8, minutes=random.randint(0, 30))
            conn.execute(text("""
                INSERT INTO access_log (employee_id, access_time, direction, gate_type, gate_name, access_result)
                VALUES (:employee_id, :access_time, 'IN', 'entry', :gate_name, 'Admitted')
            """), {
                "employee_id": emp_id,
                "access_time": check_in_time,
                "gate_name": random.choice(gate_names)
            })

            # 下班打卡 (OUT)
            check_out_time = check_in_time + timedelta(hours=9, minutes=random.randint(-30, 30))
            conn.execute(text("""
                INSERT INTO access_log (employee_id, access_time, direction, gate_type, gate_name, access_result)
                VALUES (:employee_id, :access_time, 'OUT', 'exit', :gate_name, 'success')
            """), {
                "employee_id": emp_id,
                "access_time": check_out_time,
                "gate_name": random.choice(gate_names)
            })

    print(f"✅ 每個 employee 各插入 {num_days} 天的 access_log（IN/OUT, entrance/exit）")


# === 建立 Report ===
# def insert_reports(conn, employee_ids, num_days=7):
#     for emp_id in employee_ids:
#         for i in range(num_days):
#             date = datetime.today().date() - timedelta(days=i)

#             check_in_time = datetime.combine(date, datetime.min.time()) + timedelta(hours=9, minutes=random.randint(0, 30))
#             check_out_time = check_in_time + timedelta(hours=8, minutes=random.randint(-30, 30))

#             total_hours = round((check_out_time - check_in_time).seconds / 3600, 2)
#             late_arrival = check_in_time.time() > datetime.strptime("09:00", "%H:%M").time()
#             early_departure = check_out_time.time() < datetime.strptime("18:00", "%H:%M").time()

#             conn.execute(text("""
#                 INSERT INTO report (
#                     employee_id, check_in_time, check_out_time,
#                     check_in_gate, check_out_gate, report_date,
#                     total_stay_hours, late_arrival_status, early_departure_status
#                 ) VALUES (
#                     :employee_id, :check_in_time, :check_out_time,
#                     :check_in_gate, :check_out_gate, :report_date,
#                     :total_stay_hours, :late_arrival_status, :early_departure_status
#                 )
#             """), {
#                 "employee_id": emp_id,
#                 "check_in_time": check_in_time,
#                 "check_out_time": check_out_time,
#                 "check_in_gate": "Main Entrance",
#                 "check_out_gate": "Main Entrance",
#                 "report_date": date,
#                 "total_stay_hours": total_hours,
#                 "late_arrival_status": late_arrival,
#                 "early_departure_status": early_departure
#             })
#     print(f"✅ 每個 employee 各插入 {num_days} 天的 report")

# === 主程式：自由調用 ===
if __name__ == "__main__":
    with engine.connect() as conn:
        org_ids = insert_fixed_organizations(conn)
        employee_id = insert_employees(conn, 'L1', num_employees=1, create_manager=True)
        insert_access_logs(conn, employee_id, num_days=7)
        employee_id = insert_employees(conn, 'L10', num_employees=1, create_manager=True)
        insert_access_logs(conn, employee_id, num_days=7)
        employee_id = insert_employees(conn, 'L100', num_employees=1, create_manager=True)
        insert_access_logs(conn, employee_id, num_days=7)
        employee_id = insert_employees(conn, 'L101', num_employees=1, create_manager=True)
        insert_access_logs(conn, employee_id, num_days=7)
        employee_id = insert_employees(conn, 'L11', num_employees=1, create_manager=True)
        insert_access_logs(conn, employee_id, num_days=7)
        employee_id = insert_employees(conn, 'L110', num_employees=10, create_manager=True)
        insert_access_logs(conn, employee_id, num_days=7)
        employee_id = insert_employees(conn, 'L111', num_employees=10, create_manager=True)
        insert_access_logs(conn, employee_id, num_days=7)
        employee_id = insert_employees(conn, 'L12', num_employees=1, create_manager=True)
        insert_access_logs(conn, employee_id, num_days=7)
        employee_id = insert_employees(conn, 'L120', num_employees=10, create_manager=True)
        insert_access_logs(conn, employee_id, num_days=7)
        employee_id = insert_employees(conn, 'L121', num_employees=10, create_manager=True)
        insert_access_logs(conn, employee_id, num_days=7)

        conn.commit()
    print("🎉 假資料全部插入完成！")