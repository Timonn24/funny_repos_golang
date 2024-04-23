package main

// Inspired by https://www.bytesizego.com/blog/one-billion-row-challenge-go
import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type CityTemp struct {
	min, max, sum float64
	count         int
}

type Cities map[string]CityTemp

func NewCityTemp(value float64) CityTemp {
	return CityTemp{min: value, max: value, sum: value, count: 1}
}

/*
	func (ct CityTemp) calculateAvg() float64 {
		if ct.count != 0 {
			return ct.sum / float64(ct.count)
		}
		return 0
	}
*/
func NewCities() Cities {
	return make(Cities)
}

func measureExecTime(t time.Time) {
	defer func() {
		dur := time.Since(t)
		fmt.Print("Time elapsed. ", dur)
	}()
}

func ParseFileV1(file *os.File) Cities {
	defer measureExecTime(time.Now())

	cities := NewCities()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()

		values := strings.Split(line, ";")
		if len(values) != 2 {
			continue
		}

		val, err := strconv.ParseFloat(values[1], 64)
		if err != nil {
			continue
		}

		cityName := values[0]
		if temp, ok := cities[cityName]; ok {
			temp.sum += val
			temp.count += 1
			if temp.min > val {
				temp.min = val
			}
			if temp.max < val {
				temp.max = val
			}
			cities[cityName] = temp
		} else {
			cities[cityName] = NewCityTemp(val)
		}
	}

	return cities
}

func ParseFileV2(file *os.File) Cities {
	defer measureExecTime(time.Now())

	cities := NewCities()

	producer := func() <-chan []string {
		result := make(chan []string, 0xFF)

		go func() {
			defer close(result)

			localCopy := make([]string, 0xFF)
			var count int
			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				localCopy[count] = scanner.Text()
				count++
				if count > 0xFF-1 {
					result <- localCopy
					count = 0
				}
			}
			if count != 0 {
				result <- localCopy[:count]
			}

		}()
		return result
	}

	result := producer()

	for item := range result {
		for _, text := range item {
			index := strings.Index(text, ";")
			if index == -1 {
				continue
			}

			city := text[:index]
			temp, err := strconv.ParseFloat(text[index+1:], 64)
			if err != nil {
				continue
			}

			if value, ok := cities[city]; ok {
				value.sum += temp
				value.count += 1
				if value.min > temp {
					value.min = temp
				}
				if value.max < temp {
					value.max = temp
				}
				cities[city] = value
			} else {
				cities[city] = NewCityTemp(temp)
			}
		}
	}
	return cities
}

func main() {
	file, err := os.Open("weather_stations.csv")
	if err != nil {
		panic(err)
	}
	defer file.Close()

	_ = ParseFileV2(file)
	/*
		for k, v := range cities {
			if v.count > 1 {
				fmt.Printf("City:%s, min=%f/max=%f/avg=%f\n", k, v.min, v.max, v.calculateAvg())
			}
		}
	*/
}
