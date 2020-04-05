set format y "%'.0f"
set xlabel "Days"
set ylabel "Population"

plot 	"example_result.csv" using 1:2 title "Live population", \
	"example_result.csv" using 1:3 title "Infected count", \
	"example_result.csv" using 1:4 title "Dead count", \
	"example_result.csv" using 1:5 title "In isolation", \
	"example_result.csv" using 1:6 title "Recovered"
