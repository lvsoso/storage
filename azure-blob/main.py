from azure.identity import DefaultAzureCredential
from azure.storage.blob import BlobServiceClient, BlobClient, ContainerClient


connection_string = "DefaultEndpointsProtocol=http;AccountName=devstoreaccount1;AccountKey=Eby8vdM02xNOcqFlqUwJPLlmEtlCDXJ1OUzFT50uSRZ6IFsuFq2UVErCz4I6tq/K1SZFPTOtr/KBHBeksoGMGw==;BlobEndpoint=http://127.0.0.1:10000/devstoreaccount1;"


if __name__ ==  "__main__":
    blob_service_client = BlobServiceClient.from_connection_string(connection_string)
    container_client = blob_service_client.create_container("local-test1")
    with open(file='/home/lv/lvsoso/storage/azure-blob/README.md', mode="rb") as data:
        blob_client = container_client.upload_blob(name="sample-blob.txt", data=data, overwrite=True)
    
    blob_list = container_client.list_blobs()
    for blob in blob_list:
        print(f"Name: {blob.name}")

    blob_client = blob_service_client.get_blob_client(container="local-test1", blob="sample-blob.txt")
    with open(file="/home/lv/lvsoso/storage/azure-blob/README_dl.md", mode="wb") as sample_blob:
        download_stream = blob_client.download_blob()
        sample_blob.write(download_stream.readall())        