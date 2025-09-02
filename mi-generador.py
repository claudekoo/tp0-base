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
            if "environment:" in line:
                server_lines.append(line)
                server_lines.append(f"      - NUM_AGENCIES={num_clients}\n")
            else:
                server_lines.append(line)
        if section == NETWORKS:
            networks_lines.append(line)
    
    clients_lines = []
    
    bet_data = [
        {"nombre": "Santiago Lionel", "apellido": "Lorca", "documento": "30904465", "nacimiento": "1999-03-17", "numero": "7574"},
        {"nombre": "Maria", "apellido": "Garcia", "documento": "31234567", "nacimiento": "1995-05-20", "numero": "1234"},
        {"nombre": "Juan Carlos", "apellido": "Rodriguez", "documento": "29876543", "nacimiento": "1987-11-15", "numero": "9876"}, 
        {"nombre": "Ana Lucia", "apellido": "Martinez", "documento": "32456789", "nacimiento": "2001-08-30", "numero": "5555"},
        {"nombre": "Pedro", "apellido": "Gonzalez", "documento": "28765432", "nacimiento": "1992-12-10", "numero": "7777"}
    ]
    
    for i in range(1, num_clients+1):
        client_name = f"client{i}"
        clients_lines.append(f"  {client_name}:\n")
        
        client_data = bet_data[i-1] # 5 Clients Max
        
        for line in client_lines:
            if "container_name: client1" in line:
                clients_lines.append(f"    container_name: {client_name}\n")
            elif "CLI_ID=1" in line:
                clients_lines.append(f"      - CLI_ID={i}\n")
            elif "CLI_NOMBRE=" in line:
                clients_lines.append(f"      - CLI_NOMBRE={client_data['nombre']}\n")
            elif "CLI_APELLIDO=" in line:
                clients_lines.append(f"      - CLI_APELLIDO={client_data['apellido']}\n")
            elif "CLI_DOCUMENTO=" in line:
                clients_lines.append(f"      - CLI_DOCUMENTO={client_data['documento']}\n")
            elif "CLI_NACIMIENTO=" in line:
                clients_lines.append(f"      - CLI_NACIMIENTO={client_data['nacimiento']}\n")
            elif "CLI_NUMERO=" in line:
                clients_lines.append(f"      - CLI_NUMERO={client_data['numero']}\n")
            else:
                clients_lines.append(line)
        
    with open(output_filename, "w", encoding="utf-8") as f:
        f.writelines(server_lines + clients_lines + networks_lines)

def main():
    output_filename = sys.argv[1]
    try:
        num_clients = int(sys.argv[2])
        if num_clients == 0:
            num_clients = 1
        elif num_clients < 0:
            raise ValueError("Invalid clients number")
    except ValueError as e:
        print(f"Error: {e}")
        sys.exit(1)

    generate_compose_file(output_filename, num_clients)

if __name__ == "__main__":
    main()
