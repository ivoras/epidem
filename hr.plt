set format y "%'.0f"
set xlabel "Days"
set ylabel "Population"

set yrange [0:2000]

plot 	"hr.csv" using 1:2 with linespoints title "Live population", \
	"hr.csv" using 1:3 with linespoints title "Infected count", \
	"hr.csv" using 1:4 title "Dead count", \
	"hr.csv" using 1:5 title "In isolation", \
	"hr.csv" using 1:6 with linespoints title "Recovered"
