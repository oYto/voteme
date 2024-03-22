# fetch_ticket.py
# -*- coding: utf-8 -*-

import requests
import time

def fetch_ticket():
    response = requests.post(
        "http://47.92.151.211:9090/graphql",
        json={"query": "{ getCurrentTicket { ticketID } }"}
    )
    data = response.json()
    ticket_id = data["data"]["getCurrentTicket"]["ticketID"]
    with open("current_ticket.txt", "w") as file:
        file.write(ticket_id)

if __name__ == "__main__":
    while True:
        fetch_ticket()
        time.sleep(18)  # 更新频率稍低于票据更新频率，这里假设票据每20秒更新一次

