import sys

SERVER = 0
CLIENT = 1
NETWORKS = 2

def generate_compose_file(output_filename: str, num_clients: int):
    with open("docker-compose-dev.yaml", "r", encoding="utf-8") as f:
        lines = f.readlines()
    
    server_lines = []
    client_lines = []
    networks_lines = []
    section = SERVER

    for line in lines:
        stripped = line.strip()
        if stripped.startswith("client1:"):
            section = CLIENT
            continue
        
        if section == CLIENT:
            if not line.startswith("  "):
                section = NETWORKS
                networks_lines.append(line)
                continue
            else:
                client_lines.append(line)
                continue
        if section == SERVER:
            server_lines.append(line)
        if section == NETWORKS:
            networks_lines.append(line)
    
    clients_lines = []
    for i in range(1, num_clients+1):
        client_name = f"client{i}"
        clients_lines.append(f"  {client_name}:\n")
        
        for line in client_lines:
            if "container_name: client1" in line:
                clients_lines.append(f"    container_name: {client_name}\n")
            elif "CLI_ID=1" in line:
                clients_lines.append(f"      - CLI_ID={i}\n")
            else:
                clients_lines.append(line)
        
    with open(output_filename, "w", encoding="utf-8") as f:
        f.writelines(server_lines + clients_lines + networks_lines)

def main():
    output_filename = sys.argv[1]
    try:
        num_clients = int(sys.argv[2])
        if num_clients <= 0:
            raise ValueError("Invalid clients number")
    except ValueError as e:
        print(f"Error: {e}")
        sys.exit(1)

    generate_compose_file(output_filename, num_clients)

if __name__ == "__main__":
    main()
