-- 建立 Organization 表格 (Organization_id 為 INT)
CREATE TABLE Organization (
    Organization_id VARCHAR(255) PRIMARY KEY,
    Name VARCHAR(255) NOT NULL
);

-- 建立 Employee 表格 (Employee_id 為 UUID)
CREATE TABLE Employee (
    Employee_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    First_name VARCHAR(255) NOT NULL,
    Last_name VARCHAR(255) NOT NULL,
    Is_manager BOOLEAN DEFAULT FALSE,
    Password VARCHAR(255) NOT NULL,
    Email VARCHAR(255) Not NULL,
    Organization_id VARCHAR(255),
    FOREIGN KEY (Organization_id) REFERENCES Organization(Organization_id)
);

-- 建立 Access_log 表格 (Employee_id 為 UUID)
CREATE TABLE Access_log (
    Access_id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    Employee_id UUID,
    Access_time TIMESTAMP NOT NULL,
    Direction VARCHAR(50),
    Gate_type VARCHAR(50),
    Gate_name VARCHAR(100),
    Access_result VARCHAR(50),
    FOREIGN KEY (Employee_id) REFERENCES Employee(Employee_id)
);