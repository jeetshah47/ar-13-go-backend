aws ec2 run-instances 
  --image-id ami-0c55b159cbfafe1f0   # Amazon Linux 2023
  --instance-type t3.micro 
  --key-name your-key-pair-name 
  --security-group-ids sg-xxxxxxxxx 
  --subnet-id subnet-xxxxxxxxx 
  --block-device-mappings '[{
    "DeviceName": "/dev/xvda",
    "Ebs": {
      "VolumeSize": 20,
      "VolumeType": "gp3",
      "DeleteOnTermination": true
    }
  }]' 
  --user-data file://user-data.sh 
  --tag-specifications 'ResourceType=instance,Tags=[{Key=Name,Value=ar-13-backend-free-tier}]' 
  --credit-specification CpuCredits=unlimited  # Enable unlimited mode for better performance