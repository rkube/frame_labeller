function reshape_array(data_array) {
  // data_array: a view on a 100 (= 10x10) Uint16Array
  let new_array = []
  for (let i = 0; i < 10; i++) {
    tmp_array = [data_array[i * 10], data_array[i * 10 + 1], data_array[i * 10 + 2], data_array[i * 10 + 3], data_array[i * 10 + 4], data_array[i * 10 + 5], data_array[i * 10 + 6], data_array[i * 10 + 7], data_array[i * 10 + 8], data_array[i * 10 + 9]]
    new_array.push(tmp_array)
  }
  return new_array
}
