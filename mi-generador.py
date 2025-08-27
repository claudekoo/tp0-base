import sys

def generate_compose_file(output_filename: str, num_clients: int):
    lines = []

    lines.append("name: tp0")

    lines.append("services:")

    lines.append("  server:")
    lines.append("    container_name: server")
    lines.append("    image: server:latest")
    lines.append("    entrypoint: python3 /main.py")
    lines.append("    environment:")
    lines.append("      - PYTHONUNBUFFERED=1")
    lines.append("      - LOGGING_LEVEL=DEBUG")
    lines.append("    networks:")
    lines.append("      - testing_net")

    for i in range(1, num_clients + 1):
        client_name = f"client{i}"
        lines.append(f"  {client_name}:")
        lines.append(f"    container_name: {client_name}")
        lines.append("    image: client:latest")
        lines.append("    entrypoint: /client")
        lines.append("    environment:")
        lines.append(f"      - CLI_ID={i}")
        lines.append("      - CLI_LOG_LEVEL=DEBUG")
        lines.append("    networks:")
        lines.append("      - testing_net")
        lines.append("    depends_on:")
        lines.append("      - server")

    lines.append("networks:")
    lines.append("  testing_net:")
    lines.append("    ipam:")
    lines.append("      driver: default")
    lines.append("      config:")
    lines.append("        - subnet: 172.25.125.0/24")

    with open(output_filename, "w", encoding="utf-8") as f:
        f.write("\n".join(lines) + "\n")

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
