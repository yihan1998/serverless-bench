HashOwner="c455703077a17a9b8d0fc655d939fcc6d24d819fa9a1066b74f710c35a43cbc8"
HashApp="68baea05aa0c3619b6feb78c80a07e27e4e68f921d714b8125f916c3b3370bf2"
HashFunction="c13acdc7567b225971cef2416a3a2b03c8a4d8d154df48afe75834e2f5c59ddf"
Trigger="queue"

# Check if at least one argument is provided
if [ $# -lt 3 ]; then
  echo "Usage: $0 <RuntimeInSeconds> <ReqeuestPerSecond> <AverageDurationInMilliseconds>"
  exit 1
fi

RuntimeInSeconds=$1
ReqeuestPerSecond=$2

InvocationFile="invocations.csv"

# Check if the file exists
if [ -f "$InvocationFile" ]; then
    rm "$InvocationFile"
fi

field_names="HashOwner"
field_names+=","
values=$HashOwner
values+=","

field_names+="HashApp"
field_names+=","
values+=$HashApp
values+=","

field_names+="HashFunction"
field_names+=","
values+=$HashFunction
values+=","

field_names+="Trigger"
field_names+=","
values+=$Trigger
values+=","

for i in $(seq 1 $RuntimeInSeconds); do
    if [ $i -eq $RuntimeInSeconds ]; then
        # For the last field, don't append a comma
        field_names+="$i"
        values+=$ReqeuestPerSecond
    else
        field_names+="$i,"
        values+=$ReqeuestPerSecond
        values+=","
    fi
done

# Combine field names and values into the CSV format
Content="$field_names\n$values"

echo -e $Content > $InvocationFile
cat $InvocationFile

DurationsFile="durations.csv"

# Check if the file exists
if [ -f "$DurationsFile" ]; then
    rm "$DurationsFile"
fi

range=1.0
Count=57523.0
AverageDuration=$3
Minimum=$(echo "$AverageDuration - $range" | bc)
Maximum=$(echo "$AverageDuration + $range" | bc)
percentile_Average_0=$Minimum
percentile_Average_1=$Minimum
percentile_Average_25=$Minimum
percentile_Average_50=$AverageDuration
percentile_Average_75=$Maximum
percentile_Average_99=$Maximum
percentile_Average_100=$Maximum

field_names="HashOwner,HashApp,HashFunction,Average,Count,Minimum,Maximum,percentile_Average_0,percentile_Average_1,percentile_Average_25,percentile_Average_50,percentile_Average_75,percentile_Average_99,percentile_Average_100"
values="${HashOwner},${HashApp},${HashFunction},${AverageDuration},${Count},${Minimum},${Maximum},${percentile_Average_0},${percentile_Average_1},${percentile_Average_25},${percentile_Average_50},${percentile_Average_75},${percentile_Average_99},${percentile_Average_100}"

# Combine field names and values into the CSV format
Content="$field_names\n$values"

echo -e $Content > $DurationsFile
cat $DurationsFile
