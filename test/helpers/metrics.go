package helpers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

func getMetricsEndpoint(port int) string {
	return "http://" + LocalIP() + ":" + strconv.Itoa(port) + "/metrics"
}

func httpGet(url string) ([]byte, error) {
	client := http.Client{
		Timeout: 2 * time.Second,
	}
	res, err := client.Get(url)
	if err != nil {
		fmt.Println("ERROR: could not access", url, ":", err)
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got http status code %d calling %s", res.StatusCode, url)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

func getMetricsForEndpoint(getEndpoint func() string) func() (map[string]interface{}, error) {
	return func() (map[string]interface{}, error) {
		data, err := httpGet(getEndpoint())
		if err != nil {
			return nil, err
		}

		metrics := make(map[string]interface{})
		if err := json.Unmarshal(data, &metrics); err != nil {
			return nil, err
		}

		return metrics, nil
	}
}

func GetBlockHeight(getMetrics func() (map[string]interface{}, error)) (int, error) {
	metrics, err := getMetrics()
	if err != nil {
		return 0, err
	}

	blockHeight := int(metrics["BlockStorage.BlockHeight"].(map[string]interface{})["Value"].(float64))
	fmt.Println("blockHeight", blockHeight)
	return blockHeight, nil
}

func GetMetricsForPort(httpPort int) func() (map[string]interface{}, error) {
	return getMetricsForEndpoint(func() string {
		return getMetricsEndpoint(httpPort)
	})
}
