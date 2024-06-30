package main

func compressReact(uniqueIcons map[IconDetails]int) {
	if _, isReactPresent := uniqueIcons[iconsByFileExtension["jsx"]]; !isReactPresent {
		return
	}
	delete(uniqueIcons, iconsByFileExtension["html"])
	delete(uniqueIcons, iconsByFileExtension["css"])
	delete(uniqueIcons, iconsByFileExtension["js"])
	delete(uniqueIcons, iconsByFileExtension["cjs"])
	delete(uniqueIcons, iconsByFileExtension["mjs"])
}

func compressTSX(uniqueIcons map[IconDetails]int) {
	if _, isTsxPresent := uniqueIcons[iconsByFileExtension["tsx"]]; !isTsxPresent {
		return
	}
	compressReact(uniqueIcons)
	delete(uniqueIcons, iconsByFileExtension["jsx"])
}

func compressGodot(uniqueIcons map[IconDetails]int) {
	if _, isGdPresent := uniqueIcons[iconsByFileExtension["gd"]]; !isGdPresent {
		return
	}

	delete(uniqueIcons, iconsByFileExtension["godot"])
	delete(uniqueIcons, iconsByFileExtension["tres"])
	delete(uniqueIcons, iconsByFileExtension["tscn"])
}
