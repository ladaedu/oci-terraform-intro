#! /bin/bash -eu

compartments=(
#ocid1.compartment.oc1..aaaaaaaa3sbcplfwq3y6vjsyszbxskpf6x3vxmsatasachrbau52pkmsz5wq # student-dobias
ocid1.compartment.oc1..aaaaaaaaps2cfe5cnxp4e6mjhhtz7llrbeephzxexc3uqosjbr3vwfo5ouga # student1
ocid1.compartment.oc1..aaaaaaaaotxoh7zandi2n7qaollbubhzwva4rhoy372vax3q3mlofehhcxxa # student2
ocid1.compartment.oc1..aaaaaaaa7fzjraoa7l4rlunlzbtgku2uydbuepfazv5kzmb2bcqf7tflwoda # student stepan
ocid1.compartment.oc1..aaaaaaaawopllwyltnvanzxxuakxafs6uwha3ykmnjyqiqb5da4cmfyaxrdq # student david
ocid1.compartment.oc1..aaaaaaaazdyl6y7calgvq3nkugyfxqhayq5y2fxmet7rhwzpucpekj3dwx7a # student3
ocid1.compartment.oc1..aaaaaaaaicbtnwruibyhgerp77cep5i6gnn6o6ouz74yyok4dgu2gsjhslga # student6
ocid1.compartment.oc1..aaaaaaaat7qiofmteeecxudjo64r3iawu4rnqngzqjstpqgzurp5lvcfvi2a # student7
)
profile=czechedu2021  # usually can be DEFAULT
times=(
2021-05-16T13:30Z
2021-05-17
2021-05-18
2021-05-19
2021-05-20
2021-05-21
2021-05-22
2021-05-23
2021-05-24
2021-05-25
2021-05-26
2021-05-27
2021-05-28
2021-05-29
2021-05-30
2021-05-31
)
############################################
# Usually you should not edit anything below
############################################
oci="oci --profile $profile"
times_last=$(( ${#times[@]} - 1 ))
set -x
for t in $(seq $times_last);do # seq start from 1, i.e. from 2nd element
    start_time=${times[$((t-1))]}
    end_time=${times[$t]}
    for i in ${compartments[*]};do
        out="audit-events-$i-$start_time-$end_time.log"
        if [ ! -f "$out" ];then
            $oci audit event list --start-time "$start_time" --end-time "$end_time" --compartment-id "$i" --all --skip-deserialization >> "$out"
        fi
    done
done
