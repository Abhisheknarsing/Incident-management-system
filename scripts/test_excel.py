import pandas as pd
import sys

def test_excel_file(file_path):
    """Test reading the Excel file"""
    try:
        # Read Excel file
        df = pd.read_excel(file_path)
        print("File read successfully!")
        print(f"Shape: {df.shape}")
        print("Columns:")
        for col in df.columns:
            print(f"  - {col}")
        print("\nFirst 5 rows:")
        print(df.head())
        return True
    except Exception as e:
        print(f"Error reading file: {e}")
        return False

if __name__ == "__main__":
    if len(sys.argv) != 2:
        print("Usage: python test_excel.py <file.xlsx>")
        sys.exit(1)
    
    file_path = sys.argv[1]
    success = test_excel_file(file_path)
    if not success:
        sys.exit(1)