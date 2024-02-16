# K8-Homer

K8-Homer is a dashboard service built upon [Bastien Wirtz's Homer](https://github.com/bastienwirtz/homer), tailored specifically for Kubernetes environments. This project aims to simplify the management of dashboard configurations by automating the discovery and updating process for services running within a Kubernetes cluster.

## Features

- **Auto-discovery**: K8-Homer automatically discovers new services within your Kubernetes cluster.
- **Dynamic configuration**: Configuration for discovered services is automatically updated to ensure accurate representation on the dashboard.
- **User-friendly interface**: Utilizing Homer's sleek and customizable dashboard interface, K8-Homer provides an intuitive experience for monitoring services.

## Getting Started

To get started with K8-Homer, follow these steps:

1. **Clone the Repository**:
   ```
   git clone https://github.com/robrotheram/k8-homer.git
   ```

2. **Deploy to Kubernetes**:
   ```
   kubectl apply -f k8-homer.yaml
   ```

3. **Access the Dashboard**:
Once deployed, access the K8-Homer dashboard through your browser.

## Configuration

K8-Homer leverages a `template.yml` file to manage dashboard settings. 

Only services listed in this file will have dashboard items added. Ensure that all Ingress have the following annotations to be picked up by K8-Homer:

```yaml
annotations:
    homer.item.name: Grafana
    homer.item.subtitle: Monitoring
    homer.item.logo: https://raw.githubusercontent.com/walkxcode/dashboard-icons/main/svg/grafana.svg
    homer.service.name: system #name of service to add to
```


## Configuration

K8-Homer leverages configuration files to manage dashboard settings. Customize these files to suit your preferences and requirements.

## Contributing

Contributions are welcome! If you have any ideas, improvements, or bug fixes, feel free to open an issue or submit a pull request.

## License

This project is licensed under the [MIT License](LICENSE).

## Acknowledgements

K8-Homer is built upon the foundation provided by [Bastien Wirtz's Homer](https://github.com/bastienwirtz/homer). We extend our gratitude to the contributors and maintainers of Homer for their excellent work.

## Contact

For questions or further assistance, feel free to reach out to the project maintainer at [maintainer@example.com](mailto:maintainer@example.com).