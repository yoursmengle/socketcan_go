
#include <stdio.h>
#include <stdlib.h>
#include <sys/ioctl.h>
#include <sys/socket.h>
#include <stdint.h>
#include <ctype.h>
#include <errno.h>
#include <libgen.h>
#include <signal.h>
#include <string.h>
#include <unistd.h>
#include <net/if.h>
#include <sys/types.h>
#include <linux/can/raw.h>

#define CAN_FILTER_PASS     0x01    //过滤方式-通过
#define CAN_FILTER_REJECT   0x02    //过滤方式-拒绝

__attribute__((weak))
int rcvFiltersSet(int canfd, const void *recv_ids, const uint len, const uint filterType)
{
    int i;
    if(canfd <= 0)	
        return -1;

    if(0 == recv_ids){
        setsockopt(canfd, SOL_CAN_RAW, CAN_RAW_FILTER, NULL, 0);    //不需要接收任何报文
        return 0;
    }

    struct can_filter rfilter[len];

    for(i = 0; i < len; i++){
        rfilter[i].can_id = ((uint*)recv_ids)[i];
     	if(((uint*)recv_ids)[i] & 0x80000000) {
            rfilter[i].can_mask = 0x1fffffff; 
	    } else {
		    rfilter[i].can_mask = 0x7ff;
	    }
    }

// 反向滤波暂未实现
/*    if(filterType & CAN_FILTER_REJECT){
        int join_filter = 1;
        setsockopt(canfd, SOL_CAN_RAW, CAN_RAW_JOIN_FILTERS, &join_filter, sizeof(join_filter));
    }
*/

    setsockopt(canfd, SOL_CAN_RAW, CAN_RAW_FILTER, &rfilter, sizeof(rfilter));
    return 0;
}

__attribute__((weak))
int rcvFiltersID(int canfd, const uint id, const uint len)
{
    if(canfd <= 0)	
        return -1;

    if (len < 1 || len > 8)
        return -1;

    struct can_filter rfilter;

    rfilter.can_id = id;
    rfilter.can_mask = 2^len - 1;

    setsockopt(canfd, SOL_CAN_RAW, CAN_RAW_FILTER, &rfilter, 1);
    return 0;
}