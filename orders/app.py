from flask import Flask, jsonify, request
from werkzeug.utils import secure_filename
import amaas.grpc
import os
import boto3
import json
import secrets
import logging

UPLOAD_FOLDER = '/app/uploads'

app = Flask(__name__)
app.config['UPLOAD_FOLDER'] = UPLOAD_FOLDER
app.secret_key = secrets.token_hex(32)

# Configure logging
logging.basicConfig(level=logging.INFO)

def allowed_file(filename):
    # Allow all file types
    return True

def scan_uploaded_file(file_path, handle):
    try:
        logging.info(f"Scanning file at path: {file_path}")
        logging.info(f"Handle type: {type(handle)}")

        if not isinstance(file_path, (str, bytes, os.PathLike)):
            raise TypeError("file_path must be a string, bytes, or os.PathLike object")

        result = amaas.grpc.scan_file(handle, file_path)
        
        # Log the scan result for debugging
        logging.info(f"Scan result for {file_path}: {result}")

        return result
    except TypeError as te:
        logging.error(f"TypeError during scanning: {te}")
        return None
    except Exception as e:
        logging.error(f"Error during scanning: {e}")
        return None

def upload_to_s3(file_path, bucket_name, object_name=None, content_disposition=None):
    if object_name is None:
        object_name = file_path.split("/")[-1]

    s3 = boto3.client('s3')

    try:
        extra_args = {}
        if content_disposition:
            extra_args['ContentDisposition'] = content_disposition
        extra_args['ContentType'] = "application/pdf"

        s3.upload_file(file_path, bucket_name, "downloads/"+object_name, ExtraArgs=extra_args)
        logging.info(f"File '{file_path}' uploaded to '{bucket_name}' as '{object_name}'.")
        return True
    except Exception as e:
        logging.error(f"Error uploading file to S3: {e}")
        return False

@app.route('/get-s3', methods=['GET'])
def get_s3():
    s3_url = os.environ.get('s3_object_url')
    if s3_url:
        logging.info("S3 URL found.")
        return jsonify({'s3_url': s3_url})
    else:
        logging.error("S3 URL not found.")
        return jsonify({'error': 'S3 URL not found'}), 404

@app.route('/upload', methods=['POST'])
def upload_file():
    logging.info("Received file upload request.")

    if 'file' not in request.files:
        logging.error("No file part in the request.")
        return jsonify({'error': 'No file part'}), 400

    file = request.files['file']

    if file.filename == '':
        logging.error("No file selected for upload.")
        return jsonify({'error': 'No selected file'}), 400

    if allowed_file(file.filename):
        filename = secure_filename(file.filename)
        file_path = os.path.join(app.config['UPLOAD_FOLDER'], filename)
        file.save(file_path)

        logging.info(f"File saved at {file_path}")

        # Initialize the AMAAS handle
        amaas_base_url = "antimalware.us-1.cloudone.trendmicro.com:443"
        api_key = os.environ.get('V1_API_KEY').strip('"')
        api_key = api_key.replace("ApiKey ", '')

        handle = amaas.grpc.init(amaas_base_url, api_key, True)

        # Scan the file
        scan_result = scan_uploaded_file(file_path, handle)

        if scan_result is not None:
            scan_result_dict = json.loads(scan_result)
            scan_result_code = scan_result_dict.get('scanResult', -1)

            logging.info(f"Scan Result Code {scan_result_code}")
            
            if scan_result_code == 1:
                amaas.grpc.quit(handle)
                logging.warning("File contains malware. Uploading is not allowed.")
                return jsonify({'scan_result_code': scan_result_code, 'scan_results': scan_result_dict}), 403
        else:
            amaas.grpc.quit(handle)
            logging.error("File not scanned. Check the logs for details.")
            return jsonify({'error': "File not scanned. Check the logs"}), 500

        amaas.grpc.quit(handle)

        # Upload to S3
        bucket_name = os.environ.get('s3_bucket_name')

        if upload_to_s3(file_path, bucket_name, None, 'inline'):
            return jsonify({'message': "File uploaded successfully.", 'uploaded': True, 'scan_result_code': scan_result_code, 'scan_results': scan_result_dict})
        else:
            return jsonify({'error': "Error uploading the file.", 'uploaded': False, 'scan_result_code': scan_result_code, 'scan_results': scan_result_dict}), 500

    return jsonify({'error': "File not allowed"}), 400

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)
